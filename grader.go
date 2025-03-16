package grader

import (
	"context"
	"fmt"
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
	if err := g.ClearOutputFiles(); err != nil {
		return fmt.Errorf("❌ Error clearing output files: %v", err)
	}
	files, err := os.ReadDir(g.pathOfInput)
	if err != nil {
		return fmt.Errorf("❌ Error reading input directory: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files)) // Channel เก็บ Error
	jobs := make(chan string, 10)           // จำกัดให้รันไม่เกิน 10 งานพร้อมกัน

	// 🔥 สร้าง Worker Pool (รันได้สูงสุด 10 งานพร้อมกัน)
	for i := 0; i < 10; i++ {
		go func() {
			for inputFileName := range jobs {
				if err := g.runSingleTest(fileName, inputFileName); err != nil {
					errChan <- err
				}
				wg.Done()
			}
		}()
	}

	// 🔹 ส่งไฟล์เข้า Queue (jobs)
	for _, file := range files {
		wg.Add(1)
		jobs <- file.Name()
	}
	close(jobs) // ปิด Channel หลังจากใส่งานหมดแล้ว

	wg.Wait() // รอให้ทุก Worker ทำงานเสร็จ
	close(errChan)

	// ✅ รวม Error ทั้งหมด
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
	outputFilePath := filepath.Join(g.pathOfOutput, inputFileName[:len(inputFileName)-2]+"sol")
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error opening input file %s: %v", inputFileName, err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("❌ Error reading output file %s: %v", inputFileName, err)
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
	// output, err := runCmd.CombinedOutput()
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

func (g *Grader) CheckOutput_StringMathcing() {
	files, err := os.ReadDir(g.pathOfOutput)
	if err != nil {
		fmt.Println("❌ Error reading output directory:", err)
		return
	}

	for _, file := range files {
		outputFileName := file.Name()
		outputFilePath := filepath.Join(g.pathOfOutput, outputFileName)
		expectedOutputFilePath := filepath.Join(g.pathOfExpectedOutput, outputFileName)

		outputFile, err := os.ReadFile(outputFilePath)
		if err != nil {
			fmt.Println("❌ Error reading output file:", err)
			return
		}

		expectedOutputFile, err := os.ReadFile(expectedOutputFilePath)
		if err != nil {
			fmt.Println("❌ Error reading expected output file:", err)
			return
		}

		if string(outputFile) == string(expectedOutputFile) {
			fmt.Printf("✅ Test %s passed\n", outputFileName)
		} else {
			fmt.Printf("❌ Test %s failed\n", outputFileName)
		}
	}

}

func (g *Grader) ClearOutputFiles() error {
	files, err := os.ReadDir(g.pathOfOutput)
	if err != nil {
		return fmt.Errorf("❌ Error reading output directory: %v", err)
	}

	for _, file := range files {
		filePath := filepath.Join(g.pathOfOutput, file.Name())
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("❌ Error deleting file %s: %v", filePath, err)
		}
	}

	fmt.Println("✅ All output files deleted successfully")
	return nil
}
