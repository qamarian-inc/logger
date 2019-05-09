package logger

import (
        "fmt"
        "os"
        "runtime"
        "strconv"
        "time"
)

type Logging_Info struct { // This data is needed to record logs. Functions Logger () and Log () rely on it, to operate. When creating this data, only its path should be provided, and its pointer should be what gets passed around. A single logging info should never should used for more than one Logger ().

        Log_File string // This can be a relative or absolute path to the log file. If the file doesn't exist, it is created with permission 0330.
        log_Transfer_Channel chan *log // This channel is used to transfer a log to the logger. You are not expected to set this manually. It will be set when method Run () is called.
        shutdown_Channel chan bool // This channel is used to tell loggers using this data, to shutdown.
}

type log struct { // A value of this type represents a log meant to be recorded.

        log    string // This is the actual value of the log meant to be recorded.
        state  bool  // This value tells if the log has been recorded or not. "true" means the log has been record, while "false" means the log is yet to be recorded.
        err    error // An error that occured during the logging of this log. If no error occured during the logging, this will be nil.
}

func (logging_Info *Logging_Info) Logger (log_Buffer int) (error) { // This function is a logger, and works asynchronously; in other words it is expected to run as a goroutine. This logger doesn't trigger the creation of more than 1 new OS thread when working; rather its logs are placed in a queue, and it logs them one by one, until all logs have been recorded, using a single OS thread. This way lots of logging can be made without degrading the performance of a software.

        /* To use: create a logging info, providing the path (relative or absolute) of the log file.

                Afterwards, call this function (as a gorountine). This function requires one parameter: buffer size of its log queue (the number of logs that can be queued before a caller to Log () blocks). 

                Lastly, use method Log () to add logs to the log file. */

        /* Usage Example:

                logging_Info := &Logging_Info {"log.txt"}
                go logging_Info.Logger ()
                logging_Info.Log ("Some text.") */

        defer func () {
                panic_Reason := recover ()
                if panic_Reason != nil {
                        fmt.Println ("A panic occured in the logger.")
                }
        } ()

        logging_Info.log_Transfer_Channel = make (chan *log, log_Buffer)
        logging_Info.shutdown_Channel = make (chan bool)

        // This part opens the log file (creates it, if it doesn't exist) { ...
        logFile, err0 := os.OpenFile (logging_Info.Log_File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0330)
        if err0 != nil {
                return err0
        }

        defer logFile.Close ()
        // ... }

        // Serving log requests. { ...
        for {   
                // Tries getting a new log. However, if there are no new logs and shutdown has been requested, it shutdowns. { ...
                var (
                        new_Log *log
                        value_Retrieved_Signal bool
                )

                select {
                        case new_Log, value_Retrieved_Signal = <- logging_Info.log_Transfer_Channel:
                                if value_Retrieved_Signal == false {
                                        runtime.Gosched ()
                                        continue
                                }
                        case _, _ = <- logging_Info.shutdown_Channel:
                                close (logging_Info.log_Transfer_Channel)
                                close (logging_Info.shutdown_Channel)
                                return nil
                }
                // ... }

                // This part gets the current time. Current time will be prepended to the log { ...
                currentTime := time.Now ()
                month := currentTime.Month ()
                day := currentTime.Day ()
                hour := currentTime.Hour ()
                min := currentTime.Minute ()
                // ... }

                // This part formats the log appropriately: Jan 02, 19:50: **log text**. { ...
                monthString := month.String ()
                monthString = monthString [0:3]

                var dayString string
                if day < 10 {
                        dayString = "0" + strconv.Itoa (day)
                } else {
                        dayString = strconv.Itoa (day)
                }

                var hourString string
                if hour < 10 {
                        hourString = "0" + strconv.Itoa (hour)
                } else {
                        hourString = strconv.Itoa (hour)
                }       

                var minString string
                if min < 10 {
                        minString = "0" + strconv.Itoa (min)
                } else {
                        minString = strconv.Itoa (min)
                }

                new_Log.log = fmt.Sprintf ("%s %s, %s:%s: %s\n\n", monthString, dayString, hourString, minString, new_Log.log)
                // ... }

                // Recording the log.
                _, err1 := logFile.WriteString (new_Log.log)

                if err1 != nil {
                        new_Log.err = err1
                } else {
                        new_Log.state = true
                }
        }
        // ... }

        return nil
}

func (logging_Info *Logging_Info) Log (new_Log string) (error) { // This function can be used to record a log file. Just ensure a Logger () goroutinue is running against "logging_Info" before calling this function, otherwise, this function will panic or block.

        // When a log has been successfully recorded, nil is returned. If an error wwas otherwise encountered while recording the log, the error is returned.

        log_X := log {new_Log, false, nil}

        // Sending log the logger.
        logging_Info.log_Transfer_Channel <- &log_X

        // Waits until attempt has been made to record the log. { ...
        for {
                if log_X.state == false && log_X.err == nil {
                        runtime.Gosched ()
                } else {
                        break
                }
        }
        // ... }

        if log_X.err != nil {
                return log_X.err
        } else {
                return nil
        }
}

func (logging_Info *Logging_Info) Shutdown () (error) { // This function can be used to shutdown a Logger () logging logs using this functions 'logging_Info'. If a shutdown doesn't play out appropriately, an error is returned, otherwise, nil is returned.

        // If a panic occurs it's likely due to the fact that this functions 'logging_Info' shutdown channel has been closed, or is nil. In other words, a logger is not running against this functions 'logging_Info'.
        defer func () {
                panic_Reason := recover ()
                if panic_Reason != nil {
                        fmt.Println ("A panic occured. The logger is probably not running.")
                }
        } ()

        logging_Info.shutdown_Channel <- true

        return nil
}
