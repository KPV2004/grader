package grader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Grader struct {
	pathOfSource         string
	pathOfInput          string
	pathOfOutput         string
	pathOfExpectedOutput string
}

func (g *Grader) InitGrader(pathOfSource string, pathOfInput string, pathOfOutput string, pathOfExpectedOutput string) {
	g.pathOfSource = pathOfSource
	g.pathOfInput = pathOfInput
	g.pathOfOutput = pathOfOutput
	g.pathOfExpectedOutput = pathOfExpectedOutput
}

func (g *Grader) CompileSource(fileName string, fileType string) error {
	sourceFile := filepath.Join(g.pathOfSource, fileName+"."+fileType)
	outputFile := filepath.Join(g.pathOfSource, fileName)

	compileCmd := exec.Command("g++", "-o", outputFile, sourceFile)
	if err := compileCmd.Run(); err != nil {
		return fmt.Errorf("❌ Compile error in %s: %v", fileName, err)
	}

	fmt.Printf("✅ Compile success in %s\n", fileName)
	return nil
}

func (g *Grader) RunSource(fileName string) error {
	files, err := os.ReadDir(g.pathOfInput)
	if err != nil {
		return fmt.Errorf("❌ Error reading input directory: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files)) // Buffered channel to collect errors

	for _, file := range files {
		wg.Add(1)
		go func(inputFileName string) {
			defer wg.Done()
			if err := g.runSingleTest(fileName, inputFileName); err != nil {
				errChan <- err
			}
		}(file.Name())
	}

	wg.Wait()
	close(errChan)

	// Aggregate errors
	var finalErr error
	for err := range errChan {
		if finalErr == nil {
			finalErr = err
		} else {
			finalErr = fmt.Errorf("%v\n%v", finalErr, err)
		}
	}
	return finalErr
}

func (g *Grader) runSingleTest(fileName string, inputFileName string) error {
	inputFilePath := filepath.Join(g.pathOfInput, inputFileName)
	outputFilePath := filepath.Join(g.pathOfOutput, inputFileName+".out")
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error opening input file %s: %v", inputFileName, err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error creating output file %s: %v", outputFilePath, err)
	}
	defer outputFile.Close()

	// ตั้ง Time Limit 10 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// วัดเวลาเริ่มต้น
	startTime := time.Now()

	runCmd := exec.CommandContext(ctx, filepath.Join(g.pathOfSource, fileName))
	runCmd.Stdin = inputFile
	runCmd.Stdout = outputFile

	err = runCmd.Run()
	elapsedTime := time.Since(startTime)

	// เช็คว่าโปรแกรมถูกยกเลิกเพราะ Timeout หรือไม่
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("❌ Timeout! %s took too long (>10s) with input %s", fileName, inputFileName)
	}

	if err != nil {
		return fmt.Errorf("❌ Run error in %s with input %s: %v", fileName, inputFileName, err)
	}

	fmt.Printf("✅ Run success in %s with input %s (Time: %v)\n", fileName, inputFileName, elapsedTime)
	return nil
}

func (g *Grader) CheckOutput(fileName string) error {
	files, err := os.ReadDir(g.pathOfExpectedOutput)
	if err != nil {
		return fmt.Errorf("❌ Error reading expected output directory: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files)) // Buffered channel for errors

	for _, file := range files {
		wg.Add(1)
		go func(expectedFileName string) {
			defer wg.Done()
			if err := g.compareOutput(expectedFileName); err != nil {
				errChan <- err
			}
		}(file.Name())
	}

	wg.Wait()
	close(errChan)

	// Aggregate errors
	var finalErr error
	for err := range errChan {
		if finalErr == nil {
			finalErr = err
		} else {
			finalErr = fmt.Errorf("%v\n%v", finalErr, err)
		}
	}
	return finalErr
}

func (g *Grader) compareOutput(expectedFileName string) error {
	expectedFilePath := filepath.Join(g.pathOfExpectedOutput, expectedFileName)
	outputFilePath := filepath.Join(g.pathOfOutput, expectedFileName+".out")

	expectedFile, err := os.Open(expectedFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error opening expected output file %s: %v", expectedFileName, err)
	}
	defer expectedFile.Close()

	outputFile, err := os.Open(outputFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error opening output file %s: %v", expectedFileName+".out", err)
	}
	defer outputFile.Close()

	expectedContent, err := io.ReadAll(expectedFile)
	if err != nil {
		return fmt.Errorf("❌ Error reading expected output file %s: %v", expectedFileName, err)
	}

	outputContent, err := io.ReadAll(outputFile)
	if err != nil {
		return fmt.Errorf("❌ Error reading output file %s: %v", expectedFileName+".out", err)
	}

	if !bytes.Equal(bytes.TrimSpace(expectedContent), bytes.TrimSpace(outputContent)) {
		return fmt.Errorf("❌ Output mismatch for %s", expectedFileName)
	}

	fmt.Printf("✅ Output matches for %s\n", expectedFileName)
	return nil
}
