package main

import (
	"encoding/json"
	"fmt"
)

type RoomObject struct {
	TypeID  int `json:"typeID"`
	GroupID int `json:"groupID"`
	X       int `json:"x"`
	Y       int `json:"y"`
	Z       int `json:"z"`
	Yaw     int `json:"yaw"`
	Pitch   int `json:"pitch"`
	Roll    int `json:"roll"`
	Radius  int `json:"radius"`
}

type Room struct {
	RoomName string       `json:"roomName"`
	Objects  []RoomObject `json:"objects"`
}

type Dungeon struct {
	Version     int    `json:"version"`     // Internal version for tracking of changes to dungeon "API" version
	DungeonUUID string `json:"DungeonUUID"` // Unique identifier to avoid collisions
	DungeonName string `json:"dungeonName"`
	Status      int    `json:"status"`
	FactionID   int    `json:"factionID"`
	ArchetypeID int    `json:"ArchetypeID"`
	Rooms       []Room `json:"rooms"`
}

// Simple struct used for printing out lists
type ListItem struct {
	ID   int
	Name string
}

// Get list of dungeons from database
func ListDungeons() []ListItem {
	db := getDB()
	var dungeons []ListItem

	sqlStatement := `SELECT dungeonID, dungeonName from dunDungeons`
	log.Trace("QUERY: ", sqlStatement)
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal("Failed to query db; ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry ListItem
		err = rows.Scan(&entry.ID, &entry.Name)
		dungeons = append(dungeons, entry)
	}

	db.Close()
	return dungeons
}

// Export a dungeon from the database into a JSON string
func ExportDungeon(dungeonID int) string {
	db := getDB()
	var dungeon Dungeon

	// Query dungeon
	dungeonQuery := `SELECT dungeonUUID, dungeonName, dungeonStatus, factionID, archetypeID FROM dunDungeons WHERE dungeonID = ?`
	log.Trace("QUERY: ", dungeonQuery)
	if err := db.QueryRow(dungeonQuery, dungeonID).Scan(&dungeon.DungeonUUID, &dungeon.DungeonName, &dungeon.Status, &dungeon.FactionID, &dungeon.ArchetypeID); err != nil {
		log.Fatal("Failed to query db; ", err)
	}

	// Query rooms
	roomQuery := `SELECT roomID, roomName from dunRooms WHERE dungeonID = ? ORDER BY roomID ASC`
	log.Trace("QUERY: ", roomQuery)
	if rows, err := db.Query(roomQuery, dungeonID); err != nil {
		log.Fatal("Failed to query db; ", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var room Room
			err = rows.Scan(&id, &room.RoomName)

			// Query objects
			objectQuery := `SELECT typeID, groupID, x, y, z, yaw, pitch, roll, radius FROM dunRoomObjects WHERE roomID = ?`
			log.Trace("QUERY: ", objectQuery)
			if rows, err := db.Query(objectQuery, id); err != nil {
				log.Fatal("Failed to query db; ", err)
			} else {
				defer rows.Close()
				for rows.Next() {
					var object RoomObject
					err = rows.Scan(&object.TypeID, &object.GroupID, &object.X, &object.Y, &object.Z, &object.Yaw, &object.Pitch, &object.Roll, &object.Radius)
					room.Objects = append(room.Objects, object)
				}
			}
			dungeon.Rooms = append(dungeon.Rooms, room)
		}
	}

	if output, err := json.Marshal(dungeon); err != nil {
		log.Fatal("Failed to marshal dungeon JSON; ", err)
	} else {
		return string(output)
	}
	return ""
}

// Get list of factions from database
func ListFactions() []ListItem {
	db := getDB()
	var factions []ListItem

	sqlStatement := `SELECT factionID, factionName from facFactions`
	log.Trace("QUERY: ", sqlStatement)
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal("Failed to query db; ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry ListItem
		err = rows.Scan(&entry.ID, &entry.Name)
		factions = append(factions, entry)
	}

	db.Close()
	return factions
}

// Get list of archetypes from database
func ListArchetypes() []ListItem {
	db := getDB()
	var archetypes []ListItem

	sqlStatement := `SELECT archetypeID, archetypeName from dunArchetypes`
	log.Trace("QUERY: ", sqlStatement)
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal("Failed to query db; ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry ListItem
		err = rows.Scan(&entry.ID, &entry.Name)
		archetypes = append(archetypes, entry)
	}

	db.Close()
	return archetypes
}

// Get list of rooms from database for a particular dungeon
func ListRooms(dungeonID int) []ListItem {
	db := getDB()
	var rooms []ListItem
	roomCounter := 0

	sqlStatement := `SELECT roomName from dunRooms WHERE dungeonID = ? ORDER BY roomID ASC`
	log.Trace("QUERY: ", sqlStatement)
	rows, err := db.Query(sqlStatement, dungeonID)
	if err != nil {
		log.Fatal("Failed to query db; ", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var entry ListItem
		entry.ID = roomCounter
		err = rows.Scan(&entry.Name)
		rooms = append(rooms, entry)
		roomCounter++
	}

	db.Close()
	return rooms
}

// Import a dungeon from a JSON string into the database
func ImportDungeon(data []byte, overwrite bool) int {
	db := getDB()
	var dungeon Dungeon

	if err := json.Unmarshal(data, &dungeon); err != nil {
		log.Fatal("Failed to unmarshal dungeon JSON; ", err)
		return 1
	}

	// Check if UUID is unique. If not, return an error or overwrite.
	var matchCount int
	dungeonUUIDQuery := `SELECT COUNT(*) FROM dunDungeons WHERE dungeonUUID = ?`
	log.Trace("QUERY: ", dungeonUUIDQuery)
	if err := db.QueryRow(dungeonUUIDQuery, dungeon.DungeonUUID).Scan(&matchCount); err != nil {
		log.Fatal("Failed to query db; ", err.Error())
		return 1
	} else if matchCount > 0 {
		if overwrite {
			log.Info(fmt.Sprintf("Overwriting dungeon %s.", dungeon.DungeonUUID))
		} else {
			log.Error("This dungeon already exists in the database.")
			return 1
		}
	}

	// Determine next dungeonID
	var dungeonCount int
	var dungeonID int
	dungeonCountQuery := `SELECT COUNT(*) FROM dunDungeons`
	log.Trace("QUERY: ", dungeonCountQuery)
	if err := db.QueryRow(dungeonCountQuery).Scan(&dungeonCount); err != nil {
		log.Fatal("Failed to query db; ", err.Error())
		return 1
	} else if dungeonCount == 0 {
		dungeonID = 120000000 // Default first dungeonID
	} else {
		dungeonIDQuery := `SELECT MAX(dungeonID) FROM dunDungeons`
		log.Trace("QUERY: ", dungeonIDQuery)
		if err := db.QueryRow(dungeonIDQuery).Scan(&dungeonID); err != nil {
			log.Fatal("Failed to query db; ", err.Error())
			return 1
		}
		// Increment by one to find the next valid dungeonID
		dungeonID++
	}

	// Determine next roomID
	var roomCount int
	var roomID int
	roomCountQuery := `SELECT COUNT(*) FROM dunRooms`
	log.Trace("QUERY: ", roomCountQuery)
	if err := db.QueryRow(roomCountQuery).Scan(&roomCount); err != nil {
		log.Fatal("Failed to query db; ", err.Error())
		return 1
	} else if roomCount == 0 {
		roomID = 10000 // Default first roomID
	} else {
		roomIDQuery := `SELECT MAX(roomID) FROM dunRooms`
		log.Trace("QUERY: ", roomIDQuery)
		if err := db.QueryRow(roomIDQuery).Scan(&roomID); err != nil {
			log.Fatal("Failed to query db; ", err.Error())
			return 1
		}
		//Increment by one to find the next valid roomID
		roomID++
	}

	// Determine next roomObjectID
	var roomObjectID int
	roomObjectIDQuery := `SELECT MAX(objectID) FROM dunRoomObjects`
	log.Trace("QUERY: ", roomObjectIDQuery)
	if err := db.QueryRow(roomObjectIDQuery).Scan(&roomObjectID); err != nil {
		log.Fatal("Failed to query db; ", err.Error())
		return 1
	}
	//Increment by one to find the next valid roomObjectID
	roomObjectID++

	// Insert dungeon
	dungeonQuery := `INSERT INTO dunDungeons (dungeonUUID, dungeonID, dungeonName, dungeonStatus, factionID, archetypeID) VALUES (?, ?, ?, ?, ?, ?)`
	log.Trace("QUERY: ", dungeonQuery)

	if _, err := db.Exec(dungeonQuery, dungeon.DungeonUUID, dungeonID, dungeon.DungeonName, dungeon.Status, dungeon.FactionID, dungeon.ArchetypeID); err != nil {
		log.Fatal("Failed to query db; ", err)
		return 1
	}

	// Insert rooms
	for _, room := range dungeon.Rooms {
		roomQuery := `INSERT INTO dunRooms (dungeonID, roomID, roomName) VALUES (?,?,?)`
		log.Trace("QUERY: ", roomQuery)
		if _, err := db.Exec(roomQuery, dungeonID, roomID, room.RoomName); err != nil {
			log.Fatal("Failed to query db; ", err)
			return 1
		}

		// Insert roomObjects
		roomObjectQuery := `INSERT INTO dunRoomObjects (roomID, objectID, typeID, groupID, x, y, z, yaw, pitch, roll, radius) VALUES (?,?,?,?,?,?,?,?,?,?,?)`
		log.Trace("QUERY: ", roomObjectQuery)
		for _, roomObject := range room.Objects {
			if _, err := db.Exec(roomObjectQuery, roomID, roomObjectID, roomObject.TypeID, roomObject.GroupID, roomObject.X, roomObject.Y, roomObject.Z, roomObject.Yaw, roomObject.Pitch, roomObject.Roll, roomObject.Radius); err != nil {
				log.Fatal("Failed to query db; ", err)
				return 1
			}
			roomObjectID++
		}
		roomID++
	}
	return 0
}

// Delete an entire dungeon from the database
func DeleteDungeon(dungeonID int) int {
	db := getDB()

	// Delete all room objects associated with the dungeon
	query := `DELETE FROM dunRoomObjects WHERE roomID IN (SELECT roomID FROM dunRooms WHERE dungeonID=?)`
	log.Trace("QUERY: ", query)
	if _, err := db.Exec(query, dungeonID); err != nil {
		log.Fatal("Failed to query db; ", err)
		return 1
	}

	// Delete all rooms associated with the dungeon
	query = `DELETE FROM dunRooms WHERE dungeonID=?`
	log.Trace("QUERY: ", query)
	if _, err := db.Exec(query, dungeonID); err != nil {
		log.Fatal("Failed to query db; ", err)
		return 1
	}

	// Finally, delete the dungeon itself
	query = `DELETE FROM dunDungeons WHERE dungeonID=?`
	log.Trace("QUERY: ", query)
	if _, err := db.Exec(query, dungeonID); err != nil {
		log.Fatal("Failed to query db; ", err)
		return 1
	}

	return 0
}
