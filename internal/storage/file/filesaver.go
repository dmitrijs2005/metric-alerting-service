package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

// FileSaver is a file-based implementation of the DumpSaver interface.
type FileSaver struct {
	FileStoragePath string          // Path to the dump file
	Storage         storage.Storage // Underlying metric storage
}

func (fs *FileSaver) SaveDump(ctx context.Context) error {

	x, err := fs.Storage.RetrieveAll(ctx)

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

	_, err = f.Write([]byte(dump))

	if err != nil {
		return fmt.Errorf("error writing file: %s", err.Error())
	}

	err = f.Close()

	if err != nil {
		return fmt.Errorf("error closing file: %s", err.Error())
	}

	return nil
}

func (fs *FileSaver) RestoreDump(ctx context.Context) error {

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

		parts := strings.Split(line, ":")
		metricName := parts[0]
		metricType := parts[1]
		metricValue := parts[2]

		m, err := metric.NewMetric(metric.MetricType(metricType), metricName)
		if err != nil {
			return fmt.Errorf("error creating metric: %s", err.Error())
		}

		err = fs.Storage.Add(ctx, m)
		if err != nil {
			return fmt.Errorf("error adding metric: %s", err.Error())
		}

		fs.Storage.Update(ctx, m, metricValue)

	}

	// Check for errors during scanning.
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return nil
}

func NewFileSaver(fileStoragePath string, storage storage.Storage) *FileSaver {
	return &FileSaver{fileStoragePath, storage}
}
