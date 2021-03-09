package action

import (
	"fmt"
	"github.com/rfparedes/gdg/util"
	"gopkg.in/ini.v1"
	"os"
	"os/exec"
	"strings"
)

// Gather the data
func Gather() {

	var gatherCmd *exec.Cmd
	var temp string

	cfg, err := ini.Load(util.ConfigFile)
	if err != nil {
		fmt.Printf("Failed to read config file: %v", err)
		os.Exit(1)
	}
	// Get all supported utilities
	keys := cfg.Section("utility").KeyStrings()
	// Gather for each
	for _, k := range keys {
		if strings.Contains(k, "ethtool") {
			temp = k
			k = "ethtool"
		}
		// Create dat file if it doesn't exist
		datFile := (util.DataDir + k + "/" + util.CurrentDatFile(k))
		f, err := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0744)
		util.Check(err)
		defer f.Close()
		_, err = f.WriteString(util.CreateHeader() + "\n")
		util.Check(err)
		if k == "ethtool" {
			k = temp
		}
		v := cfg.Section("utility").Key(k).Value()
		if strings.Contains(k, "ethtool") {
			_, err = f.WriteString(v + "\n")
			util.Check(err)
		}
		gatherCmd = exec.Command("bash", "-c", v)
		gatherCmd.Stdout = f
		err = gatherCmd.Start()
		util.Check(err)
		gatherCmd.Wait()
	}

}
