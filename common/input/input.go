package input

import (
	"github.com/lavalamp-/ipv666/common/config"
	"fmt"
	"github.com/lavalamp-/ipv666/common/shell"
	"log"
	"github.com/lavalamp-/ipv666/common/fs"
	"github.com/lavalamp-/ipv666/common/modeling"
	"github.com/lavalamp-/ipv666/common/addressing"
	"net"
	"github.com/lavalamp-/ipv666/common/zrandom"
	"github.com/lavalamp-/ipv666/common/statemachine"
)

func PrepareFromInputFile(inputFilePath string, fileType string, conf *config.Configuration) (error) {
	// Confirm that cleaning up is ok
	err := confirmCleanUpExisting(inputFilePath, conf)
	if err != nil {
		return err
	}
	// Load addresses from input file
	addrs, err := getIPsFromFile(inputFilePath, fileType)
	if err != nil {
		return err
	}
	// Unique addresses
	addrs = removeDuplicateIPs(addrs, conf)
	// Filter out PSLAAC addresses
	addrs = filterOutHighEntropyIPs(addrs, conf)
	// Check that enough addresses remain
	if len(addrs) < conf.InputMinAddresses {
		err := confirmTooFew(len(addrs), conf)
		if err != nil {
			return err
		}
	}
	// Delete all existing files in all directories
	err = cleanUpWorkingDirectories(conf)
	if err != nil {
		return err
	}
	// Create an empty model and write to disk
	err = createBlankModel(inputFilePath, conf)
	if err != nil {
		return err
	}
	// Write addresses to ping results file path
	writeNewAddresses(addrs, conf)
	// Update state file to indicate that ping results should be checked for blacklist
	err = updateState(conf)
	if err != nil {
		return err
	}
	return nil
}

func getIPsFromFile(inputFilePath string, inputFileType string) ([]*net.IP, error) {
	var toReturn []*net.IP
	var err error
	if inputFileType == "bin" {
		toReturn, err = addressing.ReadIPsFromBinaryFile(inputFilePath)
	} else {
		toReturn, err = addressing.ReadIPsFromHexFile(inputFilePath)
	}
	if err != nil {
		log.Printf("Error thrown when reading addresses from file '%s': %e", inputFilePath, err)
	} else {
		log.Printf("Successfully read %d addresses from %s file at '%s'.", len(toReturn), inputFileType, inputFilePath)
	}
	return toReturn, err
}

func updateState(conf *config.Configuration) (error) {
	err := statemachine.SetStateFile(conf.GetStateFilePath(), statemachine.NETWORK_GROUP)
	if err != nil {
		log.Printf("Error thrown when attempting to update state file at path '%s': %e", conf.GetStateFilePath(), err)
		return err
	}
	log.Printf("Successfully updated state file at path '%s'.", conf.GetStateFilePath())
	return nil
}

func writeNewAddresses(toWrite []*net.IP, conf *config.Configuration) (error) {
	outputPath := fs.GetTimedFilePath(conf.GetPingResultDirPath())
	log.Printf("Writing %d IP addresses to file at path '%s'.", len(toWrite), outputPath)
	err := addressing.WriteIPsToHexFile(outputPath, toWrite)
	if err != nil {
		log.Printf("Error thrown when writing addresses to path '%s': %e", outputPath, err)
		return err
	}
	log.Printf("Successfully wrote IP address list to '%s'.", outputPath)
	return nil
}

func confirmTooFew(count int, conf *config.Configuration) (error) {
	prompt := fmt.Sprintf("The resulting list of addresses is only %d long, and we recommend having at least %d to get good results. Continue? [y/N]", count, conf.InputMinAddresses)
	errMsg := fmt.Sprintf("Exiting. Please add more addresses to your input list and try again.")
	err := shell.PromptForApproval(prompt, errMsg)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func confirmCleanUpExisting(inputFilePath string, conf *config.Configuration) (error) {
	prompt := fmt.Sprintf("Provided input file at path '%s'. Starting with an input file requires cleaning up all existing state from previous runs. Continue? [y/N]", inputFilePath)
	errMsg := fmt.Sprintf("Exiting. Please backup all existing state (all directories under '%s') and try again.", conf.BaseOutputDirectory)
	err := shell.PromptForApproval(prompt, errMsg)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func filterOutHighEntropyIPs(ips []*net.IP, conf *config.Configuration) ([]*net.IP) {
	log.Printf("Now removing high entropy IP addresses from list of length %d (%f threshold, %d bits).", len(ips), conf.InputEntropyThreshold, conf.InputEntropyBitLength)
	var toReturn []*net.IP
	for i, ip := range ips {
		if i % conf.InputEmitFreq == 0 {
			log.Printf("Processing %d out of %d for high entropy IPs.", i, len(ips))
		}
		ipBytes := ([]byte)(*ip)
		entropy := zrandom.GetEntropyOfBitsFromRight(ipBytes, conf.InputEntropyBitLength)
		if entropy < conf.InputEntropyThreshold {
			toReturn = append(toReturn, ip)
		}
	}
	log.Printf("Resulting list is %d long (removed %d high entropy addresses).", len(toReturn), len(ips) - len(toReturn))
	return toReturn
}

func removeDuplicateIPs(ips []*net.IP, conf *config.Configuration) ([]*net.IP) {
	log.Printf("Now removing duplicates from list of IP addresses of length %d.", len(ips))
	toReturn := addressing.GetUniqueIPs(ips, conf.InputEmitFreq)
	log.Printf("Resulting list is %d long (removed %d duplicates).", len(toReturn), len(ips) - len(toReturn))
	return toReturn
}

func cleanUpWorkingDirectories(conf *config.Configuration) (error) {
	log.Printf("Now deleting all regular files (recursively) starting in directory '%s'.", conf.BaseOutputDirectory)
	numDeleted, err := fs.DeleteAllFilesInDirectory(conf.BaseOutputDirectory)
	if err != nil {
		log.Printf("Error thrown when deleting files under directory '%s': %e", conf.BaseOutputDirectory, err)
		return err
	}
	log.Printf("Successfully deleted %d files.", numDeleted)
	return nil
}

func createBlankModel(inputFilePath string, conf *config.Configuration) (error) {
	log.Printf("Now creating a blank statistical model.")
	model := modeling.NewAddressModel(fmt.Sprintf("Model from %s", inputFilePath))
	outputPath := fs.GetTimedFilePath(conf.GetGeneratedModelDirPath())
	log.Printf("Writing blank statistical model with name '%s' to file '%s'.", model.Name, outputPath)
	err := model.Save(outputPath)
	if err != nil {
		log.Printf("Error thrown when saving model '%s' to file '%s': %e", model.Name, outputPath, err)
		return err
	}
	return nil
}
