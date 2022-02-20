package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/psybits/gopsydose/drugdose"
)

var (
	drugname = flag.String("drug", "none", "The name of the drug."+
		"\nTry to be as accurate as possible!")
	drugroute = flag.String("route", "none", "oral, smoked, sublingual, insufflation, inhalation,"+
		"\nintravenous, etc.")
	drugargdose = flag.Float64("dose", 0, "just a number, without any units such as ml around it")
	drugunits   = flag.String("units", "none", "the units themselves: ml, L, mg etc.")
	drugperc    = flag.Float64("perc", 0, "this is only used for alcohol currently, again just a number,"+
		"\nno % around it")
	setEndTime = flag.Bool("set-end-time", false, "set the end time on the last log")
	setEndID   = flag.Int("set-end-id", 0, "set the end time on a log ID, use -get-all to get the ID")
	forUser    = flag.String("user", "default", "log for a specific user, for example if you're looking"+
		"\nafter a friend")
	getLast = flag.Int("get-last", 0, "print the last N number of logs")
	getAll  = flag.Bool("get-all", false, "print all logs")
	apiName = flag.String("apiname", "default", "the name of the API that you want to initialise for"+
		"\nsettings.toml and sources.toml")
	apiUrl = flag.String("apiurl", "default", "the URL of the API that you want to initialise for"+
		"\nsources.toml combined with -apiname")
	recreateSettings = flag.Bool("recreate-settings", false, "recreate the settings.toml file")
	recreateSources  = flag.Bool("recreate-sources", false, "recreate the sources.toml file")
	cleanLogs        = flag.Bool("clean-logs", false, "cleans the logs for the specified user name,"+
		"\noptionally using the -user option")
	removeNew = flag.Int("clean-new-logs", 0, "cleans the N number of newest logs")
	removeOld = flag.Int("clean-old-logs", 0, "cleans the N number of oldest logs")
	removeID  = flag.Int("clean-id", 0, "removes a single ID from the logs, use -get-all to see the IDs")
	dbDir     = flag.String("db-dir", "default", "full path of the DB directory, this will work only on"+
		"\nthe initial run, you can change it later in the settings.toml,"+
		"\ndon't forget to delete the old DB directory")
	getDirs       = flag.Bool("get-dirs", false, "prints the settings and DB directories path")
	localInfoDrug = flag.String("local-info-drug", "none", "print info about drug from local DB,"+
		"\nfor example if you've forgotten routes and units")
	dontLog    = flag.Bool("dont-log", false, "only fetch info about drug to local DB, but don't log anything")
	removeDrug = flag.String("remove-info-drug", "none", "remove all entries of a single drug from the info DB")
	cleanDB    = flag.Bool("clean-db", false, "remove all tables from the DB")
	getTimes   = flag.Bool("get-times", false, "get the times till onset, comeup, etc."+
		"\naccording to the current time")
	getTimesID    = flag.Int("get-times-id", 0, "get the times for a specific log id")
	stopOnCfgInit = flag.Bool("stop-on-config-init", false, "stops the program once the config"+
		"\nfiles have been initialised")
	stopOnDbInit = flag.Bool("stop-on-db-init", false, "stops the program once the DB file has"+
		"\nbeen created and initialised, if it doesn't exists already")
	verbose = flag.Bool("verbose", false, "print extra information")
)

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExample\ngopsydose -drug alcohol -route oral -dose 355 -units ml -perc 4.5")
		fmt.Fprintln(os.Stderr, "If not taken at once, when finished dosing: gopsydose -set-end-time")
		fmt.Fprintln(os.Stderr, "\n-set-end-time means, for example when finished"+
			"\ndrinking a glass of beer.")
		fmt.Fprintln(os.Stderr, "\nExample 2: gopsydose -drug lsd -route sublingual -dose 100 -units ug")
		fmt.Fprintln(os.Stderr, "The second example shouldn't require the -set-end-time command,"+
			"\nsince it's usually taken all at once.")
		fmt.Fprintln(os.Stderr, "\nTo see last dose: gopsydose -get-last 1")
		fmt.Fprintln(os.Stderr, "To see last dose progression: gopsydose -get-times")
		fmt.Fprintln(os.Stderr, "\nTo delete the 3 oldest dosages: gopsydose -clean-old-logs 3")
		fmt.Fprintln(os.Stderr, "\nCheckout the list above for more info!")
	}

	flag.Parse()

	setcfg := drugdose.InitSettingsStruct(100, *apiName, true, *dbDir, false)
	if setcfg == nil {
		fmt.Println("Settings struct not initialised.")
		os.Exit(1)
	}

	ret := setcfg.InitSettings(*recreateSettings, *verbose)
	if ret == false {
		drugdose.VerbosePrint("Settings weren't initialised.", *verbose)
	}

	gotsetcfg := drugdose.GetSettings()

	gotsrc := drugdose.InitSourceStruct(gotsetcfg.UseAPI, *apiUrl)
	ret = drugdose.InitSourceSettings(gotsrc, *recreateSources, *verbose)
	if ret == false {
		drugdose.VerbosePrint("Sources file wasn't initialised.", *verbose)
	}

	gotsrcData := drugdose.GetSourceData()

	if *getDirs == true {
		fmt.Println("DB Dir:", gotsetcfg.DBDir)
		fmt.Println("Settings Dir:", drugdose.InitSettingsDir())
	}

	if *stopOnCfgInit == true {
		os.Exit(0)
	}

	path := drugdose.CheckDBFileStruct(gotsetcfg.DBDir, "default", *verbose)
	ret = false
	if path != "" {
		ret = drugdose.CheckDBTables(path)
	}
	if path == "" || ret == false {
		if path == "" {
			path = drugdose.InitFileStructure(gotsetcfg.DBDir, "default")
		}
		ret = drugdose.InitDrugDB(gotsetcfg.UseAPI, path)
		if ret == false {
			fmt.Println("DB didn't get initialised, because of an error, exiting.")
			os.Exit(1)
		}
	}

	drugdose.VerbosePrint("Using DB: "+path, *verbose)

	if *cleanDB == true {
		ret := drugdose.CleanDB(path)
		if ret == false {
			fmt.Println("DB couldn't be cleaned, because of an error.")
		}
	}

	if *stopOnDbInit == true {
		os.Exit(0)
	}

	if *removeDrug != "none" {
		ret := drugdose.RemoveSingleDrugInfoDB(gotsetcfg.UseAPI, *removeDrug, path)
		if ret == false {
			fmt.Println("Failed to remove single drug from info DB:", *removeDrug)
		}
	}

	remAmount := 0
	revRem := false
	if *removeOld != 0 {
		remAmount = *removeOld
		revRem = true
	}

	if *removeNew != 0 {
		remAmount = *removeNew
	}

	if *cleanLogs == true || remAmount != 0 || *removeID != 0 {
		ret := drugdose.RemoveLogs(path, *forUser, remAmount, revRem, *removeID)
		if ret != true {
			fmt.Println("Couldn't remove logs because of an error.")
		}
	}

	if *getAll == true {
		ret := drugdose.GetLogs(0, *forUser, true, path, true)
		if ret == nil {
			fmt.Println("No logs could be returned.")
		}
	} else if *getLast != 0 {
		ret := drugdose.GetLogs(*getLast, *forUser, false, path, true)
		if ret == nil {
			fmt.Println("No logs could be returned.")
		}
	}

	if *localInfoDrug != "none" {
		locinfo := drugdose.GetLocalInfo(*localInfoDrug, gotsetcfg.UseAPI, path, true)
		if locinfo == nil || len(locinfo) == 0 {
			fmt.Println("Couldn't get local DB info, because of an error, for drug:", *drugname)
		}
	}

	if *setEndTime == true || *setEndID != 0 {
		ret := drugdose.SetEndTime(path, *forUser, *setEndID)
		if ret == false {
			fmt.Println("Couldn't set end time, because of an error.")
		}
	}

	if *drugname != "none" ||
		*drugroute != "none" ||
		*drugargdose != 0 ||
		*drugunits != "none" ||
		*dontLog == true {

		if *drugname == "none" {
			fmt.Println("No drug name specified, checkout: gopsydose -help")
			os.Exit(1)
		}

		if *drugroute == "none" {
			fmt.Println("No route specified, checkout: gopsydose -help")
			os.Exit(1)
		}

		if *drugargdose == 0 {
			fmt.Println("No dose specified, checkout: gopsydose -help")
			os.Exit(1)
		}

		if *drugunits == "none" {
			fmt.Println("No units specified, checkout: gopsydose -help")
			os.Exit(1)
		}

		drugdose.VerbosePrint("Using API from settings.toml: "+gotsetcfg.UseAPI, *verbose)
		drugdose.VerbosePrint("Got API URL from sources.toml: "+gotsrcData[gotsetcfg.UseAPI].ApiUrl, *verbose)

		cli := gotsetcfg.InitGraphqlClient(gotsrcData[gotsetcfg.UseAPI].ApiUrl)
		if cli != nil {
			if gotsetcfg.UseAPI == "psychonautwiki" {
				ret := gotsetcfg.FetchPsyWiki(*drugname, *drugroute, cli, path)
				if ret == false {
					fmt.Println("Didn't fetch anything.")
				}
			} else {
				fmt.Println("No valid API selected:", gotsetcfg.UseAPI)
				os.Exit(1)
			}
			if *dontLog != true {
				ret := gotsetcfg.AddToDoseDB(*forUser, *drugname, *drugroute,
					float32(*drugargdose), *drugunits, float32(*drugperc),
					path, *apiName)
				if ret == false {
					fmt.Println("Dose wasn't logged.")
				}
			}
		} else {
			fmt.Println("Something went wrong when initialising the client," +
				"\nnothing was fetched or logged.")
		}
	}

	if *getTimes == true || *getTimesID != 0 {
		times := drugdose.GetTimes(path, *forUser, gotsetcfg.UseAPI, *getTimesID, true)
		if times == nil {
			fmt.Println("Times couldn't be retrieved because of an error.")
		}
	}
}