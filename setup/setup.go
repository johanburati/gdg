package setup

import (
	"log"
	"os"
	"os/exec"

	"github.com/rfparedes/gdg/util"

	"gopkg.in/ini.v1"
)

// FindSupportedUtilities - Determine supported binaries and path
func FindSupportedUtilities() map[string]string {

	utilities := []string{"iostat", "top", "mpstat", "vmstat", "ss", "nstat", "ps", "nfsiostat", "ethtool", "ip", "pidstat", "meminfo", "slabinfo", "iofake"}
	u := make(map[string]string)

	log.Print("Finding Supported Utilities", utilities)
	for _, utility := range utilities {

		var path string
		var err error

		if utility == "meminfo" || utility == "slabinfo" {
			path, err = exec.LookPath("cat")
		} else {
			path, err = exec.LookPath(utility)
		}
		if err != nil {
			log.Printf("Cannot find %s. Excluding\n", utility)
		} else {
			u[utility] = path
		}
	}
	return u
}

// CreateOrLoadConfig - Create a configuration file
func CreateOrLoadConfig(interval string) int {

	argMap := map[string]string{
		"iostat":    " 1 3 -t -k -x -N",
		"top":       " -c -b -n 1",
		"mpstat":    " 1 2 -P ALL",
		"vmstat":    " -d",
		"ss":        " -neopa",
		"meminfo":   " /proc/meminfo",
		"slabinfo":  " /proc/slabinfo",
		"ps":        " -eo user,pid,ppid,%cpu,%mem,vsz,rss,tty,stat,start,time,wchan:32,args",
		"nfsiostat": " 1 3",
		"ethtool":   " -S",
		"ip":        " -s -s addr",
		"pidstat":   "",
		"nstat":     " -asz",
	}

	// Get current working directory to store config file and dataDir
	pwd, err := os.Getwd()
	if err != nil {
		log.Print("Cannot get current working directory")
		os.Exit(1)
	}
	configFile := pwd + "/gdg.cfg"
	dataDir := pwd + "/gdg-data/"
	log.Print(configFile)
	// Create gdg configuration file
	if err := util.CreateFile(configFile); err != nil {
		log.Println("File creation failed with error: " + err.Error())
		os.Exit(1)
	}
	// Create parent log directory
	if err := util.CreateDir(dataDir); err != nil {
		log.Println("Directory creation failed with error: " + err.Error())
		os.Exit(1)
	}

	utilities := FindSupportedUtilities()
	cfg, err := ini.Load(configFile)
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	cfg.Section("").NewKey("hostname", util.GetShortHostname())
	cfg.Section("").NewKey("interval", interval)
	cfg.Section("").NewKey("configfile", configFile)
	cfg.Section("").NewKey("datadir", dataDir)
	for u, p := range utilities {
		var call string

		//Create child log directory for utility
		if err := util.CreateDir(dataDir + u); err != nil {
			log.Println("Directory creation failed with error: " + err.Error())
			os.Exit(1)
		}
		if _, ok := argMap[u]; ok {
			call = p + argMap[u]
		} else {
			call = p
		}
		cfg.Section("utility").NewKey(u, call)
	}

	cfg.SaveTo(configFile)
	return 0
}

// Find network interfaces

// CreateSystemd function
func CreateSystemd(interval string, gdgPath string) {

	timer := `[Unit]
Description=Granular Data Gatherer Timer
Requires=gdg.service
	
[Timer]
OnActiveSec=0
OnUnitActiveSec=` + interval + "\n" +
		`AccuracySec=500msec
	
[Install]
WantedBy=timers.target`

	service := `[Unit]
Description=Granular Data Gatherer
Wants=gdg.timer
	
[Service]
Type=oneshot
ExecStart=` + gdgPath + "gdg -g\n" +
		`
[Install]
WantedBy=multi-user.target`

	strings := []string{"timer", "service"}
	// Create systemd files
	for _, s := range strings {
		f, err := os.OpenFile("/etc/systemd/system/gdg."+s, os.O_RDWR|os.O_CREATE, 0755)
		util.Check(err)
		defer f.Close()
		if s == "timer" {
			_, err := f.WriteString(timer)
			util.Check(err)
		} else {
			_, err := f.WriteString(service)
			util.Check(err)
		}
		f.Sync()
	}
}

// EnableSystemd enables the systemd gdg.timer
func EnableSystemd() {
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		log.Print("Cannot find 'systemctl' executable")
		os.Exit(1)
	}
	enableCmd := exec.Command(systemctl, "enable", "gdg.timer", "--now")
	err = enableCmd.Run()
	if err != nil {
		log.Print("Cannot enable 'gdg.timer'")
		os.Exit(1)
	}
}

// DisableSystemd disables the sytemd gdg.timer
func DisableSystemd() {

	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		log.Print("Cannot find 'systemctl' executable")
		os.Exit(1)
	}
	disableCmd := exec.Command(systemctl, "disable", "gdg.timer", "--now")
	err = disableCmd.Run()
	if err != nil {
		log.Print("Cannot disable 'gdg.timer'")
	}
}

// DeleteSystemd function to delete the gdg systemd services
func DeleteSystemd() {

	strings := []string{"timer", "service"}
	for _, s := range strings {
		err := os.Remove("/etc/systemd/system/gdg." + s)
		if err != nil {
			log.Print("Cannot remove '/etc/systemd/system/gdg." + s + "'")
		}
	}
}
