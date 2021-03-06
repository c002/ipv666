package statemachine

import (
	"github.com/lavalamp-/ipv666/common/config"
	"github.com/lavalamp-/ipv666/common/data"
	"log"
	"os"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"time"
	"bufio"
)

var addressUpdateTimer = metrics.NewTimer()

func init() {
	metrics.Register("addrupdate.file_write.time", addressUpdateTimer)
}

func updateAddressFile(conf *config.Configuration) (error) {
	cleanPings, err := data.GetCleanPingResults(conf.GetCleanPingDirPath())
	if err != nil {
		return err
	}
	//TODO don't write addresses in input file in output file
	outputPath := conf.GetOutputFilePath()
	log.Printf("Updating file at path '%s' with %d newly-found IP addresses.", outputPath, len(cleanPings))
	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	writer := bufio.NewWriter(file)
	if err != nil {
		return err
	}
	defer file.Close()
	start := time.Now()
	if conf.OutputFileType != "bin" {
		if !(conf.OutputFileType == "text") { //TODO figure out why the != check fails but this works
			log.Printf("Unexpected file format for output (%s). Defaulting to text.", conf.OutputFileType)
		}
		for _, addr := range cleanPings {
			writer.WriteString(fmt.Sprintf("%s\n", addr))
		}
	} else {
		for _, addr := range cleanPings {
			toWrite := ([]byte)(*addr)
			writer.Write(toWrite)
		}
	}
	writer.Flush()
	elapsed := time.Since(start)
	addressUpdateTimer.Update(elapsed)
	log.Printf("Finished writing %d addresses to '%s'.", len(cleanPings), outputPath)
	return nil
}
