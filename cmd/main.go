package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vsrecorder/pokemon-sprites-api/internal/controller"
	"github.com/vsrecorder/pokemon-sprites-api/internal/infrastructure"
	"github.com/vsrecorder/pokemon-sprites-api/internal/infrastructure/postgres"
	"github.com/vsrecorder/pokemon-sprites-api/internal/usecase"
	"gorm.io/gorm"
)

const (
	ExitCodeOK = iota
	ExitCodeNG
)

const (
	relativePath = "/api/v1beta"

	// listenPort は HTTP を受けるポート。docker-compose.yml の ports と揃えること。
	listenPort = "8998"

	// healthPath は死活確認の経路。認証もDBアクセスも無い。
	healthPath = "/health"

	// healthCheckArg が渡された場合はサーバを起動せず、動いているプロセスへ
	// 死活確認だけを行って終了する(docker の healthcheck から使う)。
	healthCheckArg = "--health"
)

/*
 * runHealthCheck は自分自身の /health を叩き、応答できていれば 0 を返す。
 *
 * このイメージは distroless(gcr.io/distroless/static)で、shell も curl も wget も
 * 入っていない。そのため docker の healthcheck に書ける実行ファイルは、実質この
 * バイナリ自身しかない。CMD ["/pokemon-sprites-api", "--health"] の形で使う。
 *
 * DB は見ない。ここで判定したいのは「プロセスが起きて HTTP を受けられる」ことで、
 * DB の不調まで unhealthy に含めると、DB 復旧を待つあいだ `docker compose up --wait`
 * が終わらず、コンテナ自体は正常なのにデプロイが止まってしまう。
 */
func runHealthCheck() int {
	return runHealthCheckAt(listenPort)
}

// runHealthCheckAt は指定したポートの /health を叩く。
// ポートを引数にしているのは、テストが実際に待ち受けているサービスと衝突しないよう
// 空きポートを指せるようにするため(固定ポートだと、同じ機械で本体が動いているだけで落ちる)。
func runHealthCheckAt(port string) int {
	client := &http.Client{Timeout: 3 * time.Second}

	res, err := client.Get("http://127.0.0.1:" + port + healthPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "health check failed:", err)
		return ExitCodeNG
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "health check failed: status", res.StatusCode)
		return ExitCodeNG
	}

	return ExitCodeOK
}

type APIServer struct {
	httpServer *http.Server
	db         *gorm.DB
}

func NewAPIServer(addr string, handler http.Handler, db *gorm.DB) *APIServer {
	return &APIServer{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		db: db,
	}
}

func (s *APIServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	errCh := make(chan error, 1)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil &&
			err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server error: %w", err)

	case <-ctx.Done():
		return s.Shutdown()
	}
}

func (s *APIServer) Shutdown() error {
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		return err
	}

	log.Println("cleanup: closing DB connection...")

	if sqlDB, err := s.db.DB(); err != nil {
		log.Printf("db close error: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}

	log.Println("server exited cleanly")

	return nil
}

func main() {
	// 死活確認だけを行うモード。サーバも DB 接続も要らないので最初に処理する。
	if len(os.Args) > 1 && os.Args[1] == healthCheckArg {
		os.Exit(runHealthCheck())
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	dbHostname := os.Getenv("DB_HOSTNAME")
	dbPort := os.Getenv("DB_PORT")
	userName := os.Getenv("DB_USER_NAME")
	userPassword := os.Getenv("DB_USER_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	db, err := postgres.NewDB(dbHostname, dbPort, userName, userPassword, dbName)
	if err != nil {
		log.Fatalf("failed to connect database: %v\n", err)
		os.Exit(ExitCodeNG)
	}

	r := gin.Default()

	/*
	 * 死活確認。docker の healthcheck(--health)から叩かれる。
	 *
	 * 認証もDBアクセスも持たせない。「応答できるか」だけを見る経路で、
	 * DB の状態まで混ぜると停止判定の意味が変わってしまう(runHealthCheck のコメント参照)。
	 */
	r.GET(healthPath, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "health")
	})

	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://vsrecorder.mobi",
			"https://local.vsrecorder.mobi",
		},
		AllowCredentials: false,
		MaxAge:           1 * time.Hour,
	}))

	controller.NewPokemonSprite(
		r,
		infrastructure.NewPokemonSprite(db),
		usecase.NewPokemonSprite(
			infrastructure.NewPokemonSprite(db),
		),
	).RegisterRoute(relativePath)

	{
		ctx, stop := signal.NotifyContext(
			context.Background(),
			syscall.SIGINT,
			syscall.SIGTERM,
		)
		defer stop()

		server := NewAPIServer(":"+listenPort, r, db)

		if err := server.Start(ctx); err != nil {
			log.Printf("server error: %v", err)
			os.Exit(ExitCodeNG)
		}

		os.Exit(ExitCodeOK)
	}

}
