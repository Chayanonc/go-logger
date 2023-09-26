package main

import (
	"database/sql"
	"flag"
	"fmt"
	"godb/services"
	"os"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	zsLog "github.com/rs/zerolog/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Users struct {
	Id        int
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

var db *sql.DB
var dbx *sqlx.DB

func setup() {

	zerolog.TimeFieldFormat = ""

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	}
	zsLog.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func LogV1() {
	setup()

	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	zsLog.Debug().Msg("This message appears only when log level set to Debug")
	zsLog.Info().Msg("This message appears when log level set to Debug or Info")

	now := time.Now()
	logFile := path.Join("./logs", fmt.Sprintf("%s.log", now.Format("2006-01-02")))

	_, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	if e := zsLog.Debug(); e.Enabled() {
		// Compute log output only if enabled.
		value := "bar"
		e.Str("foo", value).Msg("some debug message")
	}

}

func LogV2() {
	logFactory := services.NewLoggerFactory("./logs")

	logger, loggerClose, err := logFactory.NewLogger()
	if err != nil {
		fmt.Printf("%v", err.Error())
		os.Exit(1)
	}
	defer loggerClose()

	logger.Error("Some error",
		zap.String("key", "value"))

	logger.Info("Some data",
		zap.Error(fmt.Errorf("data")),
		zap.String("key", "value"))
}

func lowerCaseLevelEncoder(
	level zapcore.Level,
	enc zapcore.PrimitiveArrayEncoder,
) {
	if level == zap.PanicLevel || level == zap.DPanicLevel {
		enc.AppendString("error")
		return
	}

	zapcore.LowercaseLevelEncoder(level, enc)
}

func createLogger(logPath string) *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)

	now := time.Now()
	logFile := path.Join(logPath, fmt.Sprintf("%s.log", now.Format("2006-01-02")))

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})

	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	productionCfg.EncodeLevel = lowerCaseLevelEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	return zap.New(core)
}

func main() {
	var err error

	logger := createLogger("./logs")

	defer logger.Sync()

	logger.Info("Hello from Zap!")
	logger.Warn("User account is nearing the storage limit",
		zap.String("username", "john.doe"),
		zap.Float64("storageUsed", 4.5),
		zap.Float64("storageLimit", 5.0),
	)
	logger.DPanic(
		"this was never supposed to happen",
	)

	connStr := "host=localhost port=5432 user=postgres password=12341234 dbname=go sslmode=disable"
	db, err = sql.Open("postgres", connStr)

	dbx, err = sqlx.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	// ----> add user
	// user := Users{3, "Anna"}
	// if err = AddUser(user); err != nil {
	// 	panic(err)
	// }

	// ----> update user
	// user = Users{3, "Annabel"}
	// if err := UpdateUser(user); err != nil {
	// 	panic(err)
	// }

	// ----> delete user
	// if err := DeleteUser(3); err != nil {
	// 	panic(err)
	// }

	// users, err := GetUsers()

	// if err != nil {
	// 	fmt.Println("---", err)
	// 	return
	// }

	// for _, user := range users {
	// 	fmt.Println(user)
	// }

	// ----> get one user
	// user, err := GetUser(1)
	// if err != nil {
	// 	panic(err)
	// }

	// ----> get one User by dbx
	// user, err := GetUserX(1)
	// if err != nil {
	// 	println("-3-")

	// 	panic(err)
	// }

	// fmt.Println(user)
}

func GetUsersX() ([]Users, error) {
	query := "select id, name from users"
	users := []Users{}
	if err := dbx.Select(&users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserX(id int) (*Users, error) {
	query := "select id, name, created_at from users where id=$1"
	user := Users{}
	err := dbx.Get(&user, query, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
	// return nil, nil
}

// func GetUsers() ([]Users, error) {
// 	if err := db.Ping(); err != nil {
// 		return nil, err
// 	}

// 	query := "select id, name from users"
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close() // call before program close

// 	users := []Users{}
// 	for rows.Next() {
// 		user := Users{}
// 		if err = rows.Scan(&user.id, &user.name); err != nil {
// 			return nil, err
// 		}
// 		users = append(users, user)
// 	}
// 	return users, nil
// }

// func GetUser(id int) (*Users, error) {
// 	if err := db.Ping(); err != nil {
// 		return nil, err
// 	}

// 	query := "select id, name from users where id=$1"
// 	row := db.QueryRow(query, id)

// 	users := Users{}
// 	if err := row.Scan(&users.id, &users.name); err != nil {
// 		return nil, err
// 	}

// 	return &users, nil
// }

// func AddUser(user Users) error {
// 	query := "insert into users (id, name) values ($1, $2)"
// 	result, err := db.Exec(query, user.id, user.name)
// 	if err != nil {
// 		return err
// 	}

// 	affected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if affected <= 0 {
// 		return errors.New("cannot insert")
// 	}

// 	return nil
// }

// func UpdateUser(user Users) error {
// 	query := "update users set name=$1 where id=$2"
// 	result, err := db.Exec(query, user.name, user.id)
// 	if err != nil {
// 		return err
// 	}

// 	affected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if affected <= 0 {
// 		return errors.New("cannot update")
// 	}

// 	return nil
// }

// func DeleteUser(id int) error {
// 	query := "delete from users where id=$1"
// 	result, err := db.Exec(query, id)
// 	if err != nil {
// 		return err
// 	}

// 	affected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if affected <= 0 {
// 		return errors.New("cannot delete")
// 	}

// 	return nil
// }
