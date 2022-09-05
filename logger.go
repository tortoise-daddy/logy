package logy

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type LogFunction func() []interface{}

type Logger struct {
	maxFileSize int64

	fileObj *os.File

	errFileObj *os.File

	IfwFile bool

	fp string

	fn string

	Out io.Writer

	Formatter Formatter

	ReportCaller bool

	Level Level

	mu MutexWrap

	entryPool sync.Pool

	ExitFunc exitFunc

	BufferPool BufferPool
}

type exitFunc func(int)

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}


func New() *Logger {
	return &Logger{
		IfwFile: 	  false,
		Out:          os.Stderr,
		Formatter:    new(TextFormatter),
		Level:        InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
		fp:           "./",
		fn:           "app.log",
		maxFileSize:  10240,
	}
}

func (logger *Logger) newEntry() *Entry {
	entry, ok := logger.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(logger)
}

func (logger *Logger) releaseEntry(entry *Entry) {
	entry.Data = map[string]interface{}{}
	logger.entryPool.Put(entry)
}


func (logger *Logger) WithField(key string, value interface{}) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithField(key, value)
}


func (logger *Logger) WithFields(fields Fields) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithFields(fields)
}

func (logger *Logger) WithError(err error) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithError(err)
}

func (logger *Logger) zipFile(file *os.File) (*os.File, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed3,err:%v\n", err)
		return nil, err
	}
	source := fileInfo.Name()
	backLog := source + ".zip"

	zipFile, err := os.OpenFile(backLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	return nil, filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			header.Method = zip.Deflate
		}
		header.SetModTime(time.Now().UTC())
		header.Name = path
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

// Add a context to the log entry.
func (logger *Logger) WithContext(ctx context.Context) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithContext(ctx)
}

// Overrides the time of the log entry.
func (logger *Logger) WithTime(t time.Time) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithTime(t)
}

func (logger *Logger) cutFile(file *os.File) (*os.File, error) {
	// 获取当前文件名
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed2,err:%v\n", err)
		return nil, err
	}
	append := "_" + time.Now().Format("2006-01-02-15-04-05") + ".log" // 时间戳后缀
	fileName := fileInfo.Name()
	newFileName := fileName + append
	oldPath := path.Join(logger.fp, fileName) // 日志文件的全路径
	newPath := path.Join(logger.fp, newFileName)
	fmt.Println(oldPath)
	fmt.Println(newPath)
	// 关闭当前文件句柄
	logger.zipFile(file)
	file.Close()
	os.Rename(oldPath, newPath)
	fileObj, err := os.OpenFile(oldPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open logFile failed,Err:%v\n", err)
		return nil, err
	}
	return fileObj, nil
}

func (logger *Logger) checkLogSize(file *os.File) bool { // 利用os.Open()方法打开的文件句柄类型都是 *os.File 这种指针类型
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed1,err:%v\n", err)
		return false
		// panic(err)
	}
	if fileInfo.Size() > logger.maxFileSize {
		return true
	}
	return false
}
// Warning: using Log at Panic or Fatal level will not respectively Panic nor Exit.
// For this behaviour Logger.Panic or Logger.Fatal should be used instead.
func (logger *Logger) Log(level Level, args ...interface{}) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Log(level, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) LogFn(level Level, fn LogFunction) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Log(level, fn()...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Trace(args ...interface{}) {
	logger.Log(TraceLevel, args...)
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.Log(DebugLevel, args...)
}

func (logger *Logger) Info(args ...interface{}) {
	logger.Log(InfoLevel, args...)
}

func (logger *Logger) Warn(args ...interface{}) {
	logger.Log(WarnLevel, args...)
}

func (logger *Logger) Warning(args ...interface{}) {
	logger.Warn(args...)
}

func (logger *Logger) Error(args ...interface{}) {
	logger.Log(ErrorLevel, args...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.Log(FatalLevel, args...)
	logger.Exit(1)
}

func (logger *Logger) Panic(args ...interface{}) {
	logger.Log(PanicLevel, args...)
}


func (logger *Logger) Exit(code int) {
	//runHandlers()   //退出
	if logger.ExitFunc == nil {
		logger.ExitFunc = os.Exit
	}
	logger.ExitFunc(code)
}

func (logger *Logger) Setifwf(flag bool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.IfwFile = flag
}


func (logger *Logger) SetFilepn(fn, fp string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.fp = fp
	logger.fn = fn
}

func (logger *Logger) SetmaxFileSize(size int64) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.maxFileSize = size
}

func (logger *Logger) SetNoLock() {
	logger.mu.Disable()
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

// SetLevel sets the logger level.
func (logger *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

// GetLevel returns the logger level.
func (logger *Logger) GetLevel() Level {
	return logger.level()
}

// IsLevelEnabled checks if the log level of the logger is greater than the level param
func (logger *Logger) IsLevelEnabled(level Level) bool {
	return logger.level() >= level
}

// SetFormatter sets the logger formatter.
func (logger *Logger) SetFormatter(formatter Formatter) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Formatter = formatter
}

// SetOutput sets the logger output.
func (logger *Logger) SetOutput(output io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Out = output
}

func (logger *Logger) SetReportCaller(reportCaller bool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.ReportCaller = reportCaller
}

//SetBufferPool sets the logger buffer pool.
func (logger *Logger) SetBufferPool(pool BufferPool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.BufferPool = pool
}
