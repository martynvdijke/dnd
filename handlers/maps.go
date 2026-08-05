package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CampaignMap struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	Name       string `json:"name"`
	ImageURL   string `json:"image_url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	GridSize   int    `json:"grid_size"`
	IsActive   bool   `json:"is_active"`
	FogOfWar   string `json:"fog_of_war"`
}

type MapPin struct {
	ID               int64   `json:"id"`
	MapID            int64   `json:"map_id"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	X                float64 `json:"x"`
	Y                float64 `json:"y"`
	Icon             string  `json:"icon"`
	Color            string  `json:"color"`
	Description      string  `json:"description"`
	LinkedEntityType string  `json:"linked_entity_type"`
	LinkedEntityID   *int64  `json:"linked_entity_id,omitempty"`
	IsHidden         bool    `json:"is_hidden"`
	SortOrder        int     `json:"sort_order"`
}

func ListCampaignMaps(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,campaign_id,name,image_url,width,height,grid_size,is_active,fog_of_war FROM campaign_maps WHERE campaign_id=? ORDER BY name", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]CampaignMap, 0)
	for rows.Next() {
		var m CampaignMap
		rows.Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImageURL, &m.Width, &m.Height, &m.GridSize, &m.IsActive, &m.FogOfWar)
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCampaignMap(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var m CampaignMap
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO campaign_maps(campaign_id,name,image_url,width,height,grid_size,fog_of_war) VALUES(?,?,?,?,?,?,'[]')",
		campaignID, m.Name, m.ImageURL, m.Width, m.Height, m.GridSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCampaignMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var m CampaignMap
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_maps SET name=?,image_url=?,width=?,height=?,grid_size=?,is_active=?,fog_of_war=? WHERE id=?",
		m.Name, m.ImageURL, m.Width, m.Height, m.GridSize, m.IsActive, m.FogOfWar, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaignMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_map_pins WHERE map_id=?", id)
	db.DB.Exec("DELETE FROM campaign_maps WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SetActiveCampaignMap(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	mapID, _ := strconv.ParseInt(c.Param("mapId"), 10, 64)
	db.DB.Exec("UPDATE campaign_maps SET is_active=0 WHERE campaign_id=?", campaignID)
	db.DB.Exec("UPDATE campaign_maps SET is_active=1 WHERE id=?", mapID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UpdateFogOfWar(c *gin.Context) {
	mapID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		FogOfWar string `json:"fog_of_war"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_maps SET fog_of_war=? WHERE id=?", req.FogOfWar, mapID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListMapPins(c *gin.Context) {
	mapID, _ := strconv.ParseInt(c.Param("mapId"), 10, 64)
	rows, err := db.DB.Query("SELECT id,map_id,name,type,x,y,icon,color,description,linked_entity_type,linked_entity_id,is_hidden,sort_order FROM campaign_map_pins WHERE map_id=? ORDER BY sort_order,name", mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]MapPin, 0)
	for rows.Next() {
		var p MapPin
		rows.Scan(&p.ID, &p.MapID, &p.Name, &p.Type, &p.X, &p.Y, &p.Icon, &p.Color, &p.Description, &p.LinkedEntityType, &p.LinkedEntityID, &p.IsHidden, &p.SortOrder)
		out = append(out, p)
	}
	c.JSON(http.StatusOK, out)
}

func CreateMapPin(c *gin.Context) {
	mapID, _ := strconv.ParseInt(c.Param("mapId"), 10, 64)
	var p MapPin
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO campaign_map_pins(map_id,name,type,x,y,icon,color,description,linked_entity_type,linked_entity_id,is_hidden,sort_order) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
		mapID, p.Name, p.Type, p.X, p.Y, p.Icon, p.Color, p.Description, p.LinkedEntityType, p.LinkedEntityID, p.IsHidden, p.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateMapPin(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p MapPin
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_map_pins SET name=?,type=?,x=?,y=?,icon=?,color=?,description=?,linked_entity_type=?,linked_entity_id=?,is_hidden=?,sort_order=? WHERE id=?",
		p.Name, p.Type, p.X, p.Y, p.Icon, p.Color, p.Description, p.LinkedEntityType, p.LinkedEntityID, p.IsHidden, p.SortOrder, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteMapPin(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_map_pins WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetActiveCampaignMap(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var m CampaignMap
	err := db.DB.QueryRow("SELECT id,campaign_id,name,image_url,width,height,grid_size,is_active,fog_of_war FROM campaign_maps WHERE campaign_id=? AND is_active=1", campaignID).Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImageURL, &m.Width, &m.Height, &m.GridSize, &m.IsActive, &m.FogOfWar)
	if err != nil {
		// No active map, return first one
		err2 := db.DB.QueryRow("SELECT id,campaign_id,name,image_url,width,height,grid_size,is_active,fog_of_war FROM campaign_maps WHERE campaign_id=? ORDER BY id LIMIT 1", campaignID).Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImageURL, &m.Width, &m.Height, &m.GridSize, &m.IsActive, &m.FogOfWar)
		if err2 != nil {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
	}

	var fog [][]any
	json.Unmarshal([]byte(m.FogOfWar), &fog)

	pinRows, _ := db.DB.Query("SELECT id,map_id,name,type,x,y,icon,color,description,linked_entity_type,linked_entity_id,is_hidden,sort_order FROM campaign_map_pins WHERE map_id=? AND is_hidden=0 ORDER BY sort_order,name", m.ID)
	var pins []MapPin
	if pinRows != nil {
		for pinRows.Next() {
			var p MapPin
			pinRows.Scan(&p.ID, &p.MapID, &p.Name, &p.Type, &p.X, &p.Y, &p.Icon, &p.Color, &p.Description, &p.LinkedEntityType, &p.LinkedEntityID, &p.IsHidden, &p.SortOrder)
			pins = append(pins, p)
		}
		pinRows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"map":  m,
		"fog":  fog,
		"pins": pins,
	})
}
