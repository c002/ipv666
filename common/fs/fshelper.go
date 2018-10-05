package fs

import (
	"log"
	"os"
	"io/ioutil"
	"compress/zlib"
	"io"
	"bytes"
)

func CreateDirectoryIfNotExist(dirPath string) (error) {
	log.Printf("Making sure that directory at '%s' exists.", dirPath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("No directory found at path '%s'. Creating now.", dirPath)
		return os.Mkdir(dirPath, 0755)
	} else {
		log.Printf("Directory at path '%s' already exists.", dirPath)
		return nil
	}
}

func GetMostRecentFileFromDirectory(dirPath string) (string, error) {

	// https://stackoverflow.com/questions/45578172/golang-find-most-recent-file-by-date-and-time

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error thrown when reading files from directory '%s': %s", dirPath, err)
		return "", err
	}
	var newestFile string = ""
	var newestTime int64 = 0
	for _, fi := range files {
		if fi.Mode().IsRegular() {
			curTime := fi.ModTime().Unix()
			if curTime > newestTime {
				newestTime = curTime
				newestFile = fi.Name()
			}
		}
	}
	return newestFile, nil
}

func GetNonMostRecentFilesFromDirectory(dirPath string) ([]string, error) {
	var toReturn []string
	recentFile, err := GetMostRecentFileFromDirectory(dirPath)
	if err != nil || recentFile == ""{
		return toReturn, err
	}
	log.Printf("Most recent file in directory '%s' is '%s'.", dirPath, recentFile)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error thrown when trying to read files from directory '%s': '%s", dirPath, err)
		return toReturn, err
	}
	for _, fi := range files {
		name := fi.Name()
		if name != recentFile {
			toReturn = append(toReturn, name)
		}
	}
	log.Printf("Found %d files older than the most recent '%s' in directory '%s'.", len(toReturn), recentFile, dirPath)
	return toReturn, nil
}

func ZipFiles(inputPaths []string, outputPath string) (error) {
	log.Printf("Zipping up %d files (at %s) into output path of '%s'.", len(inputPaths), inputPaths, outputPath)
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error thrown when trying to create file at path '%s': %e", outFile, err)
		return err
	}
	defer outFile.Close()
	outZipFile := zlib.NewWriter(outFile)
	defer outZipFile.Close()
	for _, inputPath := range inputPaths {
		log.Printf("Now processing file at '%s'.", inputPath)
		inputFile, err := os.Open(inputPath)
		if err != nil {
			log.Printf("Error thrown when opening file at path '%s': %e", inputPath, err)
			return err
		}
		if _, err := io.Copy(outZipFile, inputFile); err != nil {
			log.Printf("Error thrown when trying to add file at '%s' to zip file at '%s': %e", inputPath, outputPath, err)
			inputFile.Close()
			return err
		}
		log.Printf("File at path '%s' successfully added to zip file at '%s'.", inputPath, outputPath)
		inputFile.Close()
	}
	log.Printf("Successfully added %d files (at %s) into output zip file at path '%s'.", len(inputPaths), inputPaths, outputPath)
	return nil
}

func CountLinesInFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return -1, err
	}
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}
	for {
		c, err := file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
