package transport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	menuApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/application"
	menuDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/domain"
)

type Handler struct {
	svc *menuApp.Service
}

func NewHandler(svc *menuApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	menus := rg.Group("/menus")

	menus.GET("", auth, requirePerm("menus:read"), h.listMenus)
	menus.POST("", auth, requirePerm("menus:write"), h.createMenu)
	menus.GET("/:id", auth, requirePerm("menus:read"), h.getMenu)
	menus.PUT("/:id", auth, requirePerm("menus:write"), h.updateMenu)
	menus.DELETE("/:id", auth, requirePerm("menus:write"), h.deleteMenu)

	menus.GET("/:id/items", auth, requirePerm("menus:read"), h.listItems)
	menus.PUT("/:id/items", auth, requirePerm("menus:write"), h.replaceItems)
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	menus := rg.Group("/menus")
	menus.GET("/:slug", h.publicGetBySlug)
}

type createMenuRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func (h *Handler) createMenu(c *gin.Context) {
	var req createMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	m, err := h.svc.CreateMenu(c.Request.Context(), req.Name, req.Slug)
	if err != nil {
		writeMenuErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, menuToJSON(m))
}

func (h *Handler) listMenus(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	items, err := h.svc.ListMenus(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]gin.H, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, menuToJSON(&m))
	}
	c.JSON(http.StatusOK, gin.H{"menus": out})
}

func (h *Handler) getMenu(c *gin.Context) {
	id := c.Param("id")
	m, err := h.svc.GetMenu(c.Request.Context(), id)
	if err != nil {
		writeMenuErr(c, err)
		return
	}
	c.JSON(http.StatusOK, menuToJSON(m))
}

type updateMenuRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) updateMenu(c *gin.Context) {
	id := c.Param("id")
	var req updateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	m, err := h.svc.UpdateMenu(c.Request.Context(), id, req.Name, req.Slug)
	if err != nil {
		writeMenuErr(c, err)
		return
	}
	c.JSON(http.StatusOK, menuToJSON(m))
}

func (h *Handler) deleteMenu(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteMenu(c.Request.Context(), id); err != nil {
		writeMenuErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) listItems(c *gin.Context) {
	menuID := c.Param("id")
	items, err := h.svc.ListItems(c.Request.Context(), menuID)
	if err != nil {
		writeMenuErr(c, err)
		return
	}

	out := make([]gin.H, 0, len(items))
	for i := range items {
		it := items[i]
		out = append(out, itemToJSON(&it))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

type replaceItemsRequest struct {
	Items []menuApp.ItemInput `json:"items" binding:"required"`
}

func (h *Handler) replaceItems(c *gin.Context) {
	menuID := c.Param("id")
	var req replaceItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.svc.ReplaceItems(c.Request.Context(), menuID, req.Items); err != nil {
		writeMenuErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) publicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	m, items, err := h.svc.PublicGetMenuBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == menuApp.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	tree := buildMenuTree(items)
	c.JSON(http.StatusOK, gin.H{
		"menu":  menuToJSON(m),
		"items": tree,
	})
}

type menuItemNode struct {
	ID        any           `json:"id"`
	ParentID  any           `json:"parent_id"`
	Label     string        `json:"label"`
	ItemType  any           `json:"item_type"`
	RefID     any           `json:"ref_id"`
	URL       any           `json:"url"`
	SortOrder int           `json:"sort_order"`
	Children  []menuItemNode `json:"children"`
}

func buildMenuTree(items []menuDomain.MenuItem) []menuItemNode {
	type node struct {
		ID        any
		ParentID  *menuDomain.MenuItemID
		Label     string
		ItemType  any
		RefID     any
		URL       any
		SortOrder int
		Children  []*node
	}

	nodes := make(map[string]*node, len(items))
	order := make([]string, 0, len(items))

	for _, it := range items {
		id := string(it.ID)
		order = append(order, id)

		var parentPtr *menuDomain.MenuItemID
		if it.ParentID != nil {
			parentPtr = it.ParentID
		}
		var ref any
		if it.RefID != nil {
			ref = *it.RefID
		}
		var url any
		if it.URL != nil {
			url = *it.URL
		}

		n := &node{
			ID:        it.ID,
			ParentID:  parentPtr,
			Label:     it.Label,
			ItemType:  it.ItemType,
			RefID:     ref,
			URL:       url,
			SortOrder: it.SortOrder,
			Children:  []*node{},
		}
		nodes[id] = n
	}

	rootPtrs := make([]*node, 0)
	for _, id := range order {
		n := nodes[id]
		if n == nil {
			continue
		}
		if n.ParentID == nil || string(*n.ParentID) == "" {
			rootPtrs = append(rootPtrs, n)
			continue
		}

		p := nodes[string(*n.ParentID)]
		if p == nil {
			rootPtrs = append(rootPtrs, n)
			continue
		}
		p.Children = append(p.Children, n)
	}

	var toJSON func(*node) menuItemNode
	toJSON = func(n *node) menuItemNode {
		var parent any
		if n.ParentID != nil {
			parent = *n.ParentID
		}

		children := make([]menuItemNode, 0, len(n.Children))
		for _, ch := range n.Children {
			children = append(children, toJSON(ch))
		}

		return menuItemNode{
			ID:        n.ID,
			ParentID:  parent,
			Label:     n.Label,
			ItemType:  n.ItemType,
			RefID:     n.RefID,
			URL:       n.URL,
			SortOrder: n.SortOrder,
			Children:  children,
		}
	}

	out := make([]menuItemNode, 0, len(rootPtrs))
	for _, r := range rootPtrs {
		out = append(out, toJSON(r))
	}
	return out
}

func writeMenuErr(c *gin.Context, err error) {
	switch err {
	case menuApp.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case menuApp.ErrAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case menuApp.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func menuToJSON(m *menuDomain.Menu) gin.H {
	return gin.H{"id": m.ID, "name": m.Name, "slug": m.Slug}
}

func itemToJSON(i *menuDomain.MenuItem) gin.H {
	var parentID any
	if i.ParentID != nil {
		parentID = *i.ParentID
	}
	var refID any
	if i.RefID != nil {
		refID = *i.RefID
	}
	var url any
	if i.URL != nil {
		url = *i.URL
	}
	return gin.H{
		"id":         i.ID,
		"menu_id":    i.MenuID,
		"parent_id":  parentID,
		"label":      i.Label,
		"item_type":  i.ItemType,
		"ref_id":     refID,
		"url":        url,
		"sort_order": i.SortOrder,
	}
}

