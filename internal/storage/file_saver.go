package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

// type FileSaver interface {
// 	Save(s Storage) error
// 	Restore(s Storage) error
// }

type FileSaver struct {
	FileStoragePath string
}

func (fs *FileSaver) SaveDump(s Storage) error {
	x, err := s.RetrieveAll()

	if err != nil {
		return err
	}

	dump := ""
	for _, m := range x {
		ms := fmt.Sprintf("%s:%s:%v", m.GetName(), m.GetType(), m.GetValue())
		dump += ms
		dump += "\n"
	}

	f, err := os.OpenFile(fs.FileStoragePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)

	if err != nil {
		return fmt.Errorf("error opening file: %s", err.Error())
	}

	fmt.Println(dump)

	n, err := f.Write([]byte(dump))

	if err != nil {
		return fmt.Errorf("error writing file: %s", err.Error())
	}

	fmt.Println(n)

	err = f.Close()

	if err != nil {
		return fmt.Errorf("error closing file: %s", err.Error())
	}

	return nil
}

func (fs *FileSaver) RestoreDump(s Storage) error {

	// Open the file for reading.
	file, err := os.Open(fs.FileStoragePath)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err.Error())
	}
	defer file.Close()

	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(file)

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text() // Get the current line.
		///fmt.Println(line)      // Process the line (here, we just print it).

		parts := strings.Split(line, ":")
		metricName := parts[0]
		metricType := parts[1]
		metricValue := parts[2]

		m, err := metric.NewMetric(metric.MetricType(metricType), metricName)
		if err != nil {
			return fmt.Errorf("error creating metric: %s", err.Error())
		}

		err = s.Add(m)
		if err != nil {
			return fmt.Errorf("error adding metric: %s", err.Error())
		}

		s.Update(m, metricValue)

	}

	// Check for errors during scanning.
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return nil
}

func NewFileSaver(fileStoragePath string) *FileSaver {
	return &FileSaver{fileStoragePath}
}
