package coffee

import (
	"github.com/kataras/iris/v12"

	"time"

	"strings"

	"github.com/XMatrixStudio/Coffee/App/controllers"
	"github.com/XMatrixStudio/Coffee/App/middleware/auth"
	"github.com/XMatrixStudio/Coffee/App/models"
	"github.com/XMatrixStudio/Coffee/App/services"
	violetSdk "github.com/XMatrixStudio/Violet.SDK.Go"
	"github.com/globalsign/mgo/bson"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
)

// Config 配置文件
type Config struct {
	Mongo   models.Mongo     `yaml:"Mongo"`   // mongoDB配置
	Server  ServerConfig     `yaml:"Server"`  // iris配置
	Violet  violetSdk.Config `yaml:"Violet"`  // Violet配置
	JWTAuth auth.JWTAuth     `yaml:"JWTAuth"` // JWTAuth
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host     string `yaml:"Host"`     // 服务器监听地址
	Port     string `yaml:"Port"`     // 服务器监听端口
	Dev      bool   `yaml:"Dev"`      // 是否开发环境
	ThumbDir string `yaml:"ThumbDir"` // 缩略图文件夹
	UserDir  string `yaml:"UserDir"`  // 用户数据文件夹
	TempDir  string `yaml:"TempDir"`  // 缓存文件夹
}

// RunServer 开始运行服务
func RunServer(c Config) {
	// 初始化数据库
	Model, err := models.NewModel(c.Mongo)
	if err != nil {
		panic(err)
	}
	// 初始化服务
	// 启动服务器
	app := iris.New()
	if c.Server.Dev {
		// app.Logger().SetLevel("debug")
	}

	switch c.JWTAuth.Store {
	case "file":
		store, err := auth.NewBuntDBStore(c.JWTAuth.FilePath)
		if err != nil {
			panic(err)
		}
		auth.Init(&c.JWTAuth, store)
		break
	}

	sessionManager := sessions.New(sessions.Config{
		Cookie:  "session_coffee_new",
		Expires: 15 * 24 * time.Hour,
	})

	Service := services.NewService(Model)

	users := mvc.New(app.Party("/user"))
	userService := Service.GetUserService()
	userService.InitViolet(c.Violet)
	users.Register(userService, sessionManager.Start)
	users.Handle(new(controllers.UsersController))

	file := mvc.New(app.Party("/file"))
	fileService := Service.GetFileService()
	fileService.InitFileService(c.Server.TempDir, c.Server.UserDir)
	file.Register(fileService, sessionManager.Start)
	file.Handle(new(controllers.FileController))

	content := mvc.New(app.Party("/content"))
	contentService := Service.GetContentService()
	contentService.SetThumbDir(c.Server.ThumbDir)
	contentService.SetUserDir(c.Server.UserDir)
	content.Register(contentService, sessionManager.Start)
	content.Handle(new(controllers.ContentController))

	app.Get("file/{fileID: string}/{filePath:string}", func(ctx iris.Context) {
		s := sessions.Get(ctx)
		fileID := ctx.Params().Get("fileID")
		filePath := ctx.Params().Get("filePath")
		if s.Get("id") == nil {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}
		if !bson.IsObjectIdHex(fileID) {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}
		filePath = strings.Replace(filePath, "|", "/", -1)
		name, err := contentService.GetFile(s.GetString("id"), fileID, filePath)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}
		ctx.SendFile(filePath, name)
		return
	})

	comment := mvc.New(app.Party("/comment"))
	comment.Register(Service.GetCommentService(), sessionManager.Start)
	comment.Handle(new(controllers.CommentController))

	like := mvc.New(app.Party("/like"))
	like.Register(Service.GetLikeService(), sessionManager.Start)
	like.Handle(new(controllers.LikeController))

	notification := mvc.New(app.Party("/notification"))
	notification.Register(Service.GetNotificationService(), sessionManager.Start)
	notification.Handle(new(controllers.NotificationController))

	app.HandleDir("/thumb", c.Server.ThumbDir)

	err = app.Run(
		// Starts the web server
		iris.Addr(c.Server.Host+":"+c.Server.Port),
		// Ignores err server closed log when CTRL/CMD+C pressed.
		iris.WithoutServerError(iris.ErrServerClosed),
		// Enables faster json serialization and more.
		iris.WithOptimizations,
	)
	if err != nil {
		panic(err)
	}
}
