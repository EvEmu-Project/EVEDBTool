package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

type SeedCommand struct {
}

func (c *SeedCommand) Help() string {
	helpText := `
Usage: evedbtool seed [options] ...
  Seeds EVEmu with default market data.
Options:
  -dryrun                Don't apply migrations, just print them.
`
	return strings.TrimSpace(helpText)
}

func (c *SeedCommand) Synopsis() string {
	return "Seeds the specified EVEmu region with default market data."
}

func (c *SeedCommand) Run(args []string) int {
	var dryrun bool

	cmdFlags := flag.NewFlagSet("up", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply query, just print it.")

	regionArray := viper.GetStringSlice("seed-regions")
	saturation := viper.GetInt("seed-saturation")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !dryrun {
		tables := GetNumberOfTables()
		if tables == 0 {
			log.Info("Database not initialized, please install the DB first")
		} else {
			log.Info("Seeding market...")
			var seedQueries []string
			for _, regionID := range regionArray {
				log.Info("Seeding region ", regionID)
				query := getQuery(regionID, saturation)
				seedQueries = append(seedQueries, query)
			}
			migrations := []*migrate.Migration{}
			migrations = append(migrations, buildSeedMigration(seedQueries))

			log.Info("Executing migration...")
			migrationSource := &migrate.MemoryMigrationSource{migrations}

			//Create a new DB connection (to avoid exhausting limit)
			db := getDB()
			migrate.SetTable("seed_migrations")
			n, err := migrate.Exec(db, "mysql", migrationSource, migrate.Up)
			if err != nil {
				log.Error("Error installing migration: ", err)
				//Check if DB died
				checkDBConnection()
			}
			db.Close()
			log.Info("Successfully applied ", n, " migration!")
		}
	} else {
		log.Info("Dry run, this is the query that will be executed:\n===\n")
		for _, regionID := range regionArray {
			query := getQuery(regionID, saturation)
			log.Info("Dry run for region ", regionID, "\n---\n")
			log.Info(query)
		}
	}

	return 0
}

func buildSeedMigration(seedQueries []string) *migrate.Migration {
	log.Debug("Building migration for seed query")

	migration := &migrate.Migration{
		Id:   "SEED_MARKET",
		Up:   seedQueries,
		Down: []string{},
	}
	return migration
}

func getQuery(regionValue string, saturation int) string {
	query := `set @regionid=RVAL; set @saturation=SVAL; create temporary table if not exists tStations (stationId int, solarSystemID int, regionID int, corporationID int, security float); truncate table tStations; select round(count(stationID)*@saturation) into @lim from staStations where regionID=@regionid ; set @i=0; insert into tStations   select stationID,solarSystemID,regionID, corporationID, security from staStations where (@i:=@i+1)<=@lim AND regionID=@regionid  order by rand(); INSERT INTO mktOrders (typeID, ownerID, regionID, stationID, price, volEntered, volRemaining, issued, minVolume, duration, solarSystemID, jumps)   SELECT typeID, corporationID, regionID, stationID, basePrice / security, 550, 550, 132478179209572976, 1, 250, solarSystemID, 1   FROM tStations, invTypes inner join invGroups USING (groupID)   WHERE invTypes.published = 1   AND invGroups.categoryID IN (4, 5, 6, 7, 8, 9, 16, 17, 18, 22, 23, 24, 25, 32, 34, 35, 39, 40, 41, 42, 43, 46); UPDATE mktOrders SET price = 100 WHERE price = 0;`
	regionMap := map[string]int{
		"Derelik":              10000001,
		"The Forge":            10000002,
		"Vale of the Silent":   10000003,
		"UUA-F4":               10000004,
		"Detorid":              10000005,
		"Wicked Creek":         10000006,
		"Cache":                10000007,
		"Scalding Pass":        10000008,
		"Insmother":            10000009,
		"Tribute":              10000010,
		"Great Wildlands":      10000011,
		"Curse":                10000012,
		"Malpais":              10000013,
		"Catch":                10000014,
		"Venal":                10000015,
		"Lonetrek":             10000016,
		"J7HZ-F":               10000017,
		"The Spire":            10000018,
		"A821-A":               10000019,
		"Tash-Murkon":          10000020,
		"Outer Passage":        10000021,
		"Stain":                10000022,
		"Pure Blind":           10000023,
		"Immensea":             10000025,
		"Etherium Reach":       10000027,
		"Molden Heath":         10000028,
		"Geminate":             10000029,
		"Heimatar":             10000030,
		"Impass":               10000031,
		"Sinq Laison":          10000032,
		"The Citadel":          10000033,
		"The Kalevala Expanse": 10000034,
		"Deklein":              10000035,
		"Devoid":               10000036,
		"Everyshore":           10000037,
		"The Bleak Lands":      10000038,
		"Esoteria":             10000039,
		"Oasa":                 10000040,
		"Syndicate":            10000041,
		"Metropolis":           10000042,
		"Domain":               10000043,
		"Solitude":             10000044,
		"Tenal":                10000045,
		"Fade":                 10000046,
		"Providence":           10000047,
		"Placid":               10000048,
		"Khanid":               10000049,
		"Querious":             10000050,
		"Cloud Ring":           10000051,
		"Kador":                10000052,
		"Cobalt Edge":          10000053,
		"Aridia":               10000054,
		"Branch":               10000055,
		"Feythabolis":          10000056,
		"Outer Ring":           10000057,
		"Fountain":             10000058,
		"Paragon Soul":         10000059,
		"Delve":                10000060,
		"Tenerifis":            10000061,
		"Omist":                10000062,
		"Period Basis":         10000063,
		"Essence":              10000064,
		"Kor-Azor":             10000065,
		"Perrigen Falls":       10000066,
		"Genesis":              10000067,
		"Verge Vendor":         10000068,
		"Black Rise":           10000069,
		"A-R00001":             11000001,
		"A-R00002":             11000002,
		"A-R00003":             11000003,
		"B-R00004":             11000004,
		"B-R00005":             11000005,
		"B-R00006":             11000006,
		"B-R00007":             11000007,
		"B-R00008":             11000008,
		"C-R00009":             11000009,
		"C-R00010":             11000010,
		"C-R00011":             11000011,
		"C-R00012":             11000012,
		"C-R00013":             11000013,
		"C-R00014":             11000014,
		"C-R00015":             11000015,
		"D-R00016":             11000016,
		"D-R00017":             11000017,
		"D-R00018":             11000018,
		"D-R00019":             11000019,
		"D-R00020":             11000020,
		"D-R00021":             11000021,
		"D-R00022":             11000022,
		"D-R00023":             11000023,
		"E-R00024":             11000024,
		"E-R00025":             11000025,
		"E-R00026":             11000026,
		"E-R00027":             11000027,
		"E-R00028":             11000028,
		"E-R00029":             11000029,
		"F-R00030":             11000030,
	}

	var regionID string

	fuzzySaturation := float32(saturation) / 100
	query_sat := strings.Replace(query, "SVAL", fmt.Sprintf("%.2f", fuzzySaturation), 1)

	if _, err := strconv.Atoi(regionValue); err == nil {
		regionID = regionValue
	} else {
		if val, ok := regionMap[regionValue]; ok {
			regionID = strconv.Itoa(val)
		} else {
			log.Error("ERR: ", regionValue, " is not a valid region name.")
			os.Exit(1)
		}
	}

	log.Info(regionID)

	return strings.Replace(query_sat, "RVAL", regionID, 1)
}
