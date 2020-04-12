package command

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/phayes/permbits"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

const COMMAND_FILE_SUFFIX = ".job"

func ScanForAvailableCommands() map[string]Command {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	paths := []string{
		fmt.Sprintf("%s/.wwmap/job.d", currentUser.HomeDir),
		"/etc/wwmap/job.d",
	}

	commands := make(map[string]Command)
	for _, p := range paths {
		log.Infof("Scan for commands in %s", p)
		found := scanForAvailableCommandsInDir(p)
		for _, c := range found {
			existing, override := commands[c.Name()]
			if override {
				log.Warn("Command %s will be replaced with %s", existing.String(), c.String())
			}
			commands[c.Name()] = c
		}
	}

	log.Infof("Detected %d commands", len(commands))

	return commands
}

func scanForAvailableCommandsInDir(path string) []Command {
	if s, err := os.Stat(path); os.IsNotExist(err) || !s.IsDir() {
		log.Warnf("%s does not exist or not a directory", path)
		return []Command{}
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("Can't list directory contents (%s): %v", path, err)
		return []Command{}
	}

	result := make([]Command, 0, len(path))
	for _, file := range files {
		if c, ok := getCommand(path, file); ok {
			log.Info("Found command: ", c)
			result = append(result, c)
		}
	}
	return result
}

func getCommand(path string, file os.FileInfo) (Command, bool) {
	if !strings.HasSuffix(file.Name(), COMMAND_FILE_SUFFIX) {
		log.Warn("File %s/%s has not suffix .job - ignored", path, file.Name())
		return nil, false
	}
	if file.IsDir() {
		log.Warn("File %s/%s is directory - ignored", path, file.Name())
		return nil, false
	}

	if !canExecute(file) {
		log.Warn("File %s/%s is not executable - ignored", path, file.Name())
		return nil, false
	}

	return FileCommand{
		name:     strings.TrimSuffix(file.Name(), COMMAND_FILE_SUFFIX),
		fullPath: filepath.Join(path, file.Name()),
	}, true
}

func canExecute(file os.FileInfo) bool {
	pb := permbits.FileMode(file.Mode())
	if pb.OtherExecute() {
		return true
	}
	fsInfo, fsOk := file.Sys().(*syscall.Stat_t)
	if !fsOk {
		return false
	}
	uid := fsInfo.Uid
	gid := fsInfo.Gid

	me, err := user.Current()
	if err != nil {
		log.Error("Can't get current user: ", err)
		return false
	}

	if fmt.Sprintf("%d", gid) == me.Gid && pb.GroupExecute() {
		return true
	}

	if fmt.Sprintf("%d", uid) == me.Uid && pb.UserExecute() {
		return true
	}

	return false
}
