package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gpayer/go-nsm/nsm"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	savePath := ""
	cID := ""
	client, err := nsm.NewClient("NSM Example Client",
		nsm.SetOptCapabilities(nsm.CapabilityClientSwitch),
		nsm.SetOpenHandler(func(projectPath, displayName, clientID string) error {
			savePath = projectPath
			cID = clientID
			return os.MkdirAll(savePath, 0755)
		}),
		nsm.SetSaveHandler(func() error {
			f, err := os.Create(path.Join(savePath, "test_save_file.txt"))
			if err != nil {
				return err
			}
			_, err = f.WriteString("Hello, World!\n")
			if err != nil {
				return err
			}
			f.WriteString("ClientID: ")
			f.WriteString(cID)
			f.WriteString("\n")
			return nil
		}))
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	for {
		select {
		case <-signals:
			return
		case <-time.After(time.Second):
		}
		if client.State == nsm.StateError {
			fmt.Printf("ERROR: %v\n", client.Error)
			return
		}
	}
}
