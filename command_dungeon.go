package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

// Dungeon CLI root command
type DungeonCommand struct {
}

func (c *DungeonCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon [options] ...
  Manage EVEmu dungeons.
Subcommands:
  list                   List all available dungeons in the database.
  apply                  Apply all dungeons from the dungeon directory.
  import                 Import a dungeon from file.
  export                 Export a dungeon to file.
  new                    Creates a new blank dungeon.
  delete                 Delete a dungeon from the database.
  add-room               Add a new room to a dungeon.
  remove-room            Remove a room from a dungeon.
  list-rooms             Lists all rooms in the specified dungeon.
  list-archetypes        List available dungeon archetypes.
  list-factions          List available factions.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonCommand) Synopsis() string {
	return "Manages EVEmu dungeons."
}

func (c *DungeonCommand) Run(args []string) int {
	fmt.Println(c.Help())

	return 0
}

// List dungeons
type DungeonListCommand struct {
}

func (c *DungeonListCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon list [options] ...
  Lists all dungeons in the database.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonListCommand) Synopsis() string {
	return "Lists all dungeons in the database."
}

func (c *DungeonListCommand) Run(args []string) int {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Dungeon ID", "Dungeon Name"})
	table.SetColWidth(60)

	listOutput := ListDungeons()

	for _, entry := range listOutput {
		table.Append([]string{
			strconv.Itoa(entry.ID),
			entry.Name,
		})
	}

	table.Render()
	return 0
}

// Export a dungeon to file
type DungeonExportCommand struct {
}

func (c *DungeonExportCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon export [options] ...
  Exports a specific dungeon from the database to a file.
  Options:
  -dungeon               ID of the dungeon to export.
  -file                  Filename to write the dungeon data to.
  -dryrun                Don't write the output to file, just print it.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonExportCommand) Synopsis() string {
	return "Exports a specific dungeon from the database to a file."
}

func (c *DungeonExportCommand) Run(args []string) int {
	var dungeonID int
	var outputFilename string
	var dryrun bool

	cmdFlags := flag.NewFlagSet("export", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&dungeonID, "dungeon", 0, "ID of the dungeon to export.")
	cmdFlags.StringVar(&outputFilename, "file", "export.json", "Filename to write the dungeon data to.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't write the output to file, just print it.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	if dungeonID < 1 {
		ui.Output(c.Help())
		return 1
	}

	jsonOutput := ExportDungeon(dungeonID)

	if dryrun {
		fmt.Println(jsonOutput)
	} else {
		if _, err := os.Stat(outputFilename); err == nil {
			log.Error("File already exists, not overwriting.")
			return 1
		} else {
			if err := ioutil.WriteFile(outputFilename, []byte(jsonOutput), 0644); err != nil {
				log.Error("Error writing file: ", err)
				return 1
			} else {
				log.Info("Successfully exported dungeon to: ", outputFilename)
			}
		}
	}

	return 0
}

// Import a dungeon from file
type DungeonImportCommand struct {
}

func (c *DungeonImportCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon import <filename> [options] ...
  Imports a dungeon from provided JSON file.
  Options:
  -file <filename>       File to import.
  -overwrite             Overwrite existing dungeon if a match is found.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonImportCommand) Synopsis() string {
	return "Imports a dungeon from provided JSON file."
}

func (c *DungeonImportCommand) Run(args []string) int {
	var filename string
	var overwrite bool

	cmdFlags := flag.NewFlagSet("import", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.StringVar(&filename, "file", "", "File to import.")
	cmdFlags.BoolVar(&overwrite, "overwrite", false, "Overwrite existing dungeon if a match is found.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	if filename == "" {
		ui.Output(c.Help())
		return 1
	}

	if _, err := os.Stat(filename); err == nil {
		if data, err := ioutil.ReadFile(filename); err != nil {
			log.Error("Error reading file: ", err)
			return 1
		} else {
			ImportDungeon(data, overwrite)
			log.Info("Successfully imported dungeon!")
		}
	} else {
		log.Error("File does not exist: ", filename)
		return 1
	}

	return 0
}

// Apply all dungeons from dungeons directory
type DungeonApplyCommand struct {
}

func (c *DungeonApplyCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon apply [options] ...
  Applies all dungeons from the dungeon directory to the database.
  Options:
  -overwrite             Overwrite existing dungeon if a match is found.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonApplyCommand) Synopsis() string {
	return "Applies all dungeons from the dungeon directory to the database."
}

func (c *DungeonApplyCommand) Run(args []string) int {
	var overwrite bool

	cmdFlags := flag.NewFlagSet("apply", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.BoolVar(&overwrite, "overwrite", false, "Overwrite existing dungeon if a match is found.")

	items, _ := ioutil.ReadDir(viper.GetString("dungeon-dir"))
	log.Info(fmt.Sprintf("Attempting to import %d dungeons...", len(items)))
	successCount := 0
	for _, item := range items {
		if !item.IsDir() {
			fullpath := filepath.Join(viper.GetString("dungeon-dir"), item.Name())
			log.Trace("Import candidate: ", fullpath)
			if _, err := os.Stat(fullpath); err == nil {
				if data, err := ioutil.ReadFile(fullpath); err != nil {
					log.Error("Error reading file: ", err)
					return 1
				} else {
					ImportDungeon(data, overwrite)
					successCount++
				}
			}
		}
	}
	log.Info(fmt.Sprintf("Successfully imported %d dungeons!", successCount))
	return 0
}

// Create a new blank dungeon
type DungeonNewCommand struct {
}

func (c *DungeonNewCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon new [options] ...
  Creates a new blank dungeon in the database.
  Options:
  -name <string>         Name of the new dungeon.
  -status <1-3>          Status (1=Release, 2=Testing, 3=Working Copy).
  -faction <int>         Faction ID (run 'evedbtool dungeon list-factions' to list available)
  -archetype <int>       Archetype ID (run 'evedbtool dungeon list-archetypes' to list available)
  -dryrun                Don't create dungeon, just print the JSON data.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonNewCommand) Synopsis() string {
	return "Creates a new blank dungeon in the database."
}

func (c *DungeonNewCommand) Run(args []string) int {
	var dungeon Dungeon
	var err error
	var dryrun bool

	cmdFlags := flag.NewFlagSet("new", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.StringVar(&dungeon.DungeonName, "name", "", "Name of the new dungeon.")
	cmdFlags.IntVar(&dungeon.Status, "status", 0, "Status (1=Release, 2=Testing, 3=Working Copy).")
	cmdFlags.IntVar(&dungeon.FactionID, "faction", 0, "Faction ID (run 'evedbtool dungeon list-factions' to list available)")
	cmdFlags.IntVar(&dungeon.ArchetypeID, "archetype", 0, "Archetype ID (run 'evedbtool dungeon list-archetypes' to list available)")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't create dungeon, just print the JSON data.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	// Interactively ask for dungeon name
	if dungeon.DungeonName == "" {
		dungeon.DungeonName = StringPrompt("Dungeon Name: ")
	}

	// Interactively ask for dungeon status
	if dungeon.Status == 0 {
		dungeon.Status, err = strconv.Atoi(StringPrompt("Status: (1=Release, 2=Testing, 3=Working Copy) "))
		if err != nil {
			log.Error("Error parsing response: ", err)
			return 1
		}
	}

	// Interactively ask for dungeon faction
	if dungeon.FactionID == 0 {
		input := StringPrompt("Faction ID: (Type L to list all available) ")
		if input == "L" {

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Faction ID", "Faction Name"})
			table.SetColWidth(60)

			listOutput := ListFactions()

			for _, entry := range listOutput {
				table.Append([]string{
					strconv.Itoa(entry.ID),
					entry.Name,
				})
			}

			table.Render()

			input = StringPrompt("Faction ID: ")
		}

		dungeon.FactionID, err = strconv.Atoi(input)
		if err != nil {
			log.Error("Error parsing response: ", err)
			return 1
		}
	}

	// Interactively ask for dungeon archetype
	if dungeon.ArchetypeID == 0 {
		input := StringPrompt("Archetype ID: (Type L to list all available) ")
		if input == "L" {

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Archetype ID", "Archetype Name"})
			table.SetColWidth(60)

			listOutput := ListArchetypes()

			for _, entry := range listOutput {
				table.Append([]string{
					strconv.Itoa(entry.ID),
					entry.Name,
				})
			}

			table.Render()

			input = StringPrompt("Archetype ID: ")
		}

		dungeon.ArchetypeID, err = strconv.Atoi(input)
		if err != nil {
			log.Error("Error parsing response: ", err)
			return 1
		}

	}

	// Generate a uuid for this dungeon
	dungeon.DungeonUUID = uuid.New().String()

	// Create or print the dungeon
	if data, err := json.Marshal(dungeon); err != nil {
		log.Error("Error marshalling new dungeon: ", err)
	} else {
		if dryrun {
			fmt.Println(string(data))
			return 0
		} else {
			ImportDungeon(data, false)
		}
	}

	return 0
}

// List all factions
type DungeonFactionListCommand struct {
}

func (c *DungeonFactionListCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon list-factions ...
  Lists all faction IDs along with their names.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonFactionListCommand) Synopsis() string {
	return "Lists all faction IDs along with their names."
}

func (c *DungeonFactionListCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("list-factions", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Faction ID", "Faction Name"})
	table.SetColWidth(60)

	listOutput := ListFactions()

	for _, entry := range listOutput {
		table.Append([]string{
			strconv.Itoa(entry.ID),
			entry.Name,
		})
	}

	table.Render()
	return 0
}

// List all Archetypes
type DungeonArchetypeListCommand struct {
}

func (c *DungeonArchetypeListCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon list-archetypes ...
  Lists all Archetype IDs along with their names.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonArchetypeListCommand) Synopsis() string {
	return "Lists all Archetype IDs along with their names."
}

func (c *DungeonArchetypeListCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("list-archetypes", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Archetype ID", "Archetype Name"})
	table.SetColWidth(60)

	listOutput := ListArchetypes()

	for _, entry := range listOutput {
		table.Append([]string{
			strconv.Itoa(entry.ID),
			entry.Name,
		})
	}

	table.Render()
	return 0
}

// Delete a dungeon from the database
type DungeonDeleteCommand struct {
}

func (c *DungeonDeleteCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon delete [options] ...
  Deletes a dungeon from the database.
  Options:
  -dungeon <int>                ID of the dungeon to delete.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonDeleteCommand) Synopsis() string {
	return "Deletes a dungeon from the database."
}

func (c *DungeonDeleteCommand) Run(args []string) int {
	var dungeonID int

	cmdFlags := flag.NewFlagSet("delete", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&dungeonID, "dungeon", 0, "ID of the dungeon to delete.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	if dungeonID == 0 {
		ui.Output(c.Help())
		return 1
	}

	if DeleteDungeon(dungeonID) == 0 {
		log.Info("Successfully deleted the dungeon.")
	}

	return 0
}

// Adds a room to the specified dungeon in the database.
type DungeonAddRoomCommand struct {
}

func (c *DungeonAddRoomCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon add-room [options] ...
Adds a room to the specified dungeon in the database.
  Options:
  -dungeon <int>         ID of the dungeon for which to add the room.
  -name <string>         Name of the room to add.
  -dryrun                Don't apply change, just print the JSON output.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonAddRoomCommand) Synopsis() string {
	return "Adds a room to the specified dungeon in the database."
}

func (c *DungeonAddRoomCommand) Run(args []string) int {
	var dungeon Dungeon
	var room Room
	var dryrun bool
	var dungeonID int

	cmdFlags := flag.NewFlagSet("new", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&dungeonID, "dungeon", 0, "ID of the dungeon for which to add the room.")
	cmdFlags.StringVar(&room.RoomName, "name", "", "Name of the room to add.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply change, just print the JSON output.")

	if err := cmdFlags.Parse(args); err != nil {
		ui.Output(c.Help())
		return 1
	}

	if dungeonID == 0 {
		ui.Output(c.Help())
		return 1
	}

	// Get dungeon from database
	if err := json.Unmarshal([]byte(ExportDungeon(dungeonID)), &dungeon); err != nil {
		log.Error("Error unmarshalling")
		return 1
	}

	// Interactively ask for room name
	if room.RoomName == "" {
		room.RoomName = StringPrompt("Room Name: ")
	}

	// Add room to dungeon
	dungeon.Rooms = append(dungeon.Rooms, room)

	// Create or print the dungeon
	if data, err := json.Marshal(dungeon); err != nil {
		log.Error("Error marshalling new dungeon: ", err)
	} else {
		if dryrun {
			fmt.Println("Updated dungeon: ")
			fmt.Println(string(data))
			return 0
		} else {
			if DeleteDungeon(dungeonID) == 0 {
				ImportDungeon(data, false)
			} else {
				log.Error("Error deleting existing dungeon, not importing updated dungeon.")
			}
		}
	}
	return 0
}

// Removes the specified room from the database.
type DungeonRemoveRoomCommand struct {
}

func (c *DungeonRemoveRoomCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon remove-room [options] ...
  Removes the specified room from a dungeon in the database.
  Options:
  -dungeon <int>         ID of the dungeon for which to remove the room.
  -room <int>            ID of the room to remove.
  -dryrun                Don't apply change, just print the JSON output.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonRemoveRoomCommand) Synopsis() string {
	return "Removes the specified room from a dungeon in the database."
}

func (c *DungeonRemoveRoomCommand) Run(args []string) int {
	var dungeonID int
	var roomID int
	var dryrun bool

	cmdFlags := flag.NewFlagSet("remove-room", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&dungeonID, "dungeon", 0, "ID of the dungeon for which to remove the room.")
	cmdFlags.IntVar(&roomID, "room", 99, "ID of the room to remove.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply change, just print the JSON output.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	if dungeonID < 1 {
		ui.Output(c.Help())
		return 1
	}
	if roomID == 99 {
		ui.Output(c.Help())
		return 1
	}

	var dungeon Dungeon

	if err := json.Unmarshal([]byte(ExportDungeon(dungeonID)), &dungeon); err != nil {
		log.Error("Error unmarshalling")
		return 1
	}

	// Update slice
	dungeon.Rooms = append(dungeon.Rooms[:roomID], dungeon.Rooms[roomID+1:]...)

	// Create or print the dungeon
	if data, err := json.Marshal(dungeon); err != nil {
		log.Error("Error marshalling new dungeon: ", err)
	} else {
		if dryrun {
			fmt.Println("Updated dungeon: ")
			fmt.Println(string(data))
			return 0
		} else {
			if DeleteDungeon(dungeonID) == 0 {
				ImportDungeon(data, false)
			} else {
				log.Error("Error deleting existing dungeon, not importing updated dungeon.")
			}
		}
	}

	return 0
}

// Lists all rooms in the specified dungeon.
type DungeonRoomListCommand struct {
}

func (c *DungeonRoomListCommand) Help() string {
	helpText := `
Usage: evedbtool dungeon list-rooms ...
    Lists all rooms in the specified dungeon.
	Options:
	-dungeon <int>         ID of the dungeon for which to add the room.
`
	return strings.TrimSpace(helpText)
}

func (c *DungeonRoomListCommand) Synopsis() string {
	return "Lists all rooms in the specified dungeon."
}

func (c *DungeonRoomListCommand) Run(args []string) int {
	var dungeonID int

	cmdFlags := flag.NewFlagSet("list-rooms", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&dungeonID, "dungeon", 0, "ID of the dungeon to export.")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error("Error parsing arguments")
		return 1
	}

	if dungeonID < 1 {
		ui.Output(c.Help())
		return 1
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Room ID", "Room Name"})
	table.SetColWidth(60)

	listOutput := ListRooms(dungeonID)

	if len(listOutput) == 0 {
		log.Error("Either dungeon does not exist, or has no rooms.")
		return 1
	}

	for _, entry := range listOutput {
		table.Append([]string{
			strconv.Itoa(entry.ID),
			entry.Name,
		})
	}

	table.Render()
	return 0
}
