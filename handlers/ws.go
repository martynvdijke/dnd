package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"villum/db"
	"villum/middleware"
)

const (
	wsReadLimit     = 4096
	wsReadDeadline  = 60 * time.Second
	wsPingInterval  = 30 * time.Second
	wsWriteDeadline = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSClient struct {
	hub    *WSHub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	role   string
}

type WSHub struct {
	mu      sync.RWMutex
	clients map[int64][]*WSClient
}

var Hub = &WSHub{
	clients: make(map[int64][]*WSClient),
}

func (h *WSHub) Register(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.userID] = append(h.clients[client.userID], client)
}

func (h *WSHub) Unregister(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.clients[client.userID]
	for i, c := range clients {
		if c == client {
			h.clients[client.userID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *WSHub) BroadcastToUser(userID int64, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients[userID] {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func (h *WSHub) BroadcastToAdmins(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.clients {
		for _, c := range clients {
			if c.role == "admin" {
				select {
				case c.send <- msg:
				default:
				}
			}
		}
	}
}

func (h *WSHub) BroadcastToCampaignMembers(campaignID int64, msg []byte) {
	var memberIDs []int64
	rows, err := db.DB.Query("SELECT user_id FROM campaign_members WHERE campaign_id=?", campaignID)
	if err != nil {
		middleware.LogWarn("ws", "campaign members query failed", "error", err)
	} else {
		func() {
			defer rows.Close()
			for rows.Next() {
				var uid int64
				rows.Scan(&uid)
				memberIDs = append(memberIDs, uid)
			}
			if err := rows.Err(); err != nil {
				middleware.LogWarn("ws", "rows iteration failed", "error", err)
			}
		}()
	}
	var ownerID int64
	db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", campaignID).Scan(&ownerID)
	if ownerID > 0 {
		memberIDs = append(memberIDs, ownerID)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range memberIDs {
		for _, c := range h.clients[uid] {
			select {
			case c.send <- msg:
			default:
			}
		}
	}
}

func SendCharacterUpdate(charID int64) {
	payload, _ := json.Marshal(map[string]int64{"character_id": charID})
	msg, _ := json.Marshal(WSMessage{Type: "character_update", Payload: payload})
	Hub.BroadcastToAdmins(msg)

	var ownerID int64
	db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", charID).Scan(&ownerID)
	if ownerID > 0 {
		Hub.BroadcastToUser(ownerID, msg)
	}
	var campaignID *int64
	db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id=?", charID).Scan(&campaignID)
	if campaignID != nil {
		Hub.BroadcastToCampaignMembers(*campaignID, msg)
	}
}

func SendPartyUpdate() {
	msg, _ := json.Marshal(WSMessage{Type: "party_update", Payload: json.RawMessage(`{}`)})
	Hub.BroadcastToAdmins(msg)
}

// Live table event types (campaign-scoped, server -> client).
const (
	WSEventDiceRoll        = "dice_roll"
	WSEventCombatUpdate    = "combat_update"
	WSEventKnowledgeReveal = "knowledge_reveal"
	WSEventSceneEffect     = "scene_effect"
	WSEventSpellCast       = "spell_cast"
	WSEventAmbience        = "ambience"
)

// broadcastToCampaign marshals a payload and fans it out to campaign members.
func broadcastToCampaign(campaignID int64, eventType string, payload any) {
	if campaignID <= 0 {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg, _ := json.Marshal(WSMessage{Type: eventType, Payload: b})
	Hub.BroadcastToCampaignMembers(campaignID, msg)
}

// DiceRollEvent is the dice_roll payload for the live table.
type DiceRollEvent struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	CharacterID int64  `json:"character_id"`
	Expression  string `json:"expression"`
	Total       int    `json:"total"`
	Text        string `json:"text"`
}

// SendDiceRoll relays a persisted roll to the character's campaign members.
func SendDiceRoll(campaignID, userID, characterID int64, username, expression string, total int, text string) {
	broadcastToCampaign(campaignID, WSEventDiceRoll, DiceRollEvent{
		UserID:      userID,
		Username:    username,
		CharacterID: characterID,
		Expression:  expression,
		Total:       total,
		Text:        text,
	})
}

// SendCombatUpdate signals members to refetch the combat view for a campaign.
func SendCombatUpdate(campaignID int64) {
	broadcastToCampaign(campaignID, WSEventCombatUpdate, map[string]int64{"campaign_id": campaignID})
}

// SceneEffectEvent is the scene_effect payload played on the live table.
type SceneEffectEvent struct {
	SceneID int64  `json:"scene_id"`
	Effect  string `json:"effect"`
}

// SendSceneEffect relays a scene's special effect to campaign members.
func SendSceneEffect(campaignID, sceneID int64, effect string) {
	broadcastToCampaign(campaignID, WSEventSceneEffect, SceneEffectEvent{SceneID: sceneID, Effect: effect})
}

// SpellCastEvent is the spell_cast payload for the live table.
type SpellCastEvent struct {
	CharacterID int64  `json:"character_id"`
	SpellID     int64  `json:"spell_id"`
	Name        string `json:"name"`
	School      string `json:"school"`
	Level       int    `json:"level"`
	Effect      string `json:"effect"`
}

// SendSpellCast relays a cast spell to campaign members (banner + WLED).
func SendSpellCast(campaignID, characterID, spellID int64, name, school string, level int, effect string) {
	broadcastToCampaign(campaignID, WSEventSpellCast, SpellCastEvent{
		CharacterID: characterID,
		SpellID:     spellID,
		Name:        name,
		School:      school,
		Level:       level,
		Effect:      effect,
	})
}

// AmbienceEvent tells clients to play or stop a synthesized ambience loop.
type AmbienceEvent struct {
	Track  string  `json:"track"`
	Action string  `json:"action"`
	Volume float64 `json:"volume"`
}

// SendAmbience relays an ambience play/stop command to campaign members.
func SendAmbience(campaignID int64, track, action string, volume float64) {
	broadcastToCampaign(campaignID, WSEventAmbience, AmbienceEvent{Track: track, Action: action, Volume: volume})
}

// SendKnowledgeReveal broadcasts a knowledge entry that just became shared.
func SendKnowledgeReveal(entry any) {
	k, ok := entry.(Knowledge)
	if !ok || !k.Shared {
		return
	}
	broadcastToCampaign(k.CampaignID, WSEventKnowledgeReveal, map[string]any{
		"id":          k.ID,
		"campaign_id": k.CampaignID,
		"title":       k.Title,
		"content":     k.Content,
		"source":      k.Source,
		"status":      k.Status,
		"shared":      k.Shared,
	})
}

func HandleWebSocket(c *gin.Context) {
	sessionID, err := c.Cookie("session")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	sess := middleware.Store.Get(sessionID)
	if sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		middleware.LogWarn("ws", "websocket upgrade error", "error", err)
		return
	}

	client := &WSClient{
		hub:    Hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: sess.UserID,
		role:   sess.Role,
	}
	Hub.Register(client)

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(wsReadLimit)
	c.conn.SetReadDeadline(time.Now().Add(wsReadDeadline))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(wsReadDeadline))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(wsPingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteDeadline))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteDeadline))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
