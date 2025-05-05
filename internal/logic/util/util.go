package util

import (
	"context"
	"fmt"
	"gf_cms/internal/service"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"math"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// Util

type sUtil struct{}

var (
	insUtil           = sUtil{}
	Ctx               context.Context
	ProjectName       *gvar.Var
	BackendPrefix     *gvar.Var
	SystemRoot        string
	BackendGroup      string
	BackendApiGroup   string
	PcApiGroup        string
	MobileApiGroup    string
	PublicCachePreFix string
	ServerRoot        *gvar.Var
	DryRun            bool
)

func init() {
	service.RegisterUtil(New())
	Ctx = gctx.New()
	//项目ProjectName
	ProjectName, _ = g.Cfg().Get(Ctx, "server.projectName", "gf_cms")
	//后台入口前缀
	BackendPrefix, _ = g.Config().Get(Ctx, "server.backendPrefix", "backend")
	//BackendGroup 后台view分组
	BackendGroup = "/" + BackendPrefix.String()
	//BackendApiGroup 后台api分组
	BackendApiGroup = "/" + BackendPrefix.String() + "_api"
	PcApiGroup = "/api"
	MobileApiGroup = "/mobile_api"
	//公共缓存前缀
	PublicCachePreFix = ProjectName.String() + ":public"
	//项目目录
	SystemRoot, _ = os.Getwd()
	//服务目录
	ServerRoot, _ = g.Cfg().Get(Ctx, "server.serverRoot")
	//空跑模式
	dryRun, _ := g.Cfg().Get(Ctx, "database.default.dryRun")
	DryRun = gconv.Bool(dryRun)
}

func New() *sUtil {
	return &sUtil{}
}

func Util() *sUtil {
	return &insUtil
}

// DryRun 获取空跑模式
func (*sUtil) DryRun() bool {
	return DryRun
}

// ProjectName 获取ProjectName
func (*sUtil) ProjectName() string {
	return ProjectName.String()
}

// JwtKey 获取JwtKey
func (*sUtil) JwtKey() string {
	jwtKey := Util().GetConfig("server.jwtKey")
	if jwtKey == "" {
		return ProjectName.String()
	}
	return jwtKey
}

// SystemRoot 获取SystemRoot
func (*sUtil) SystemRoot() string {
	return SystemRoot
}

// BackendPrefix 后台入口前缀
func (*sUtil) BackendPrefix() string {
	return BackendPrefix.String()
}

// BackendApiPrefix 后台入口前缀
func (*sUtil) BackendApiPrefix() string {
	return service.Util().BackendPrefix() + "_api"
}

// BackendGroup 后台view分组
func (*sUtil) BackendGroup() string {
	return "/" + Util().BackendPrefix()
}

// BackendApiGroup 后台api分组
func (*sUtil) BackendApiGroup() string {
	return "/" + Util().BackendPrefix() + "_api"
}

// PcApiGroup pcApi分组
func (*sUtil) PcApiGroup() string {
	return PcApiGroup
}

// MApiGroup 移动Api分组
func (*sUtil) MApiGroup() string {
	return MobileApiGroup
}

// ServerRoot 服务目录
func (s *sUtil) ServerRoot() string {
	return ServerRoot.String()
}

// GetConfig 获取配置文件的配置信息
func (*sUtil) GetConfig(node string) string {
	config, _ := g.Cfg().Get(Ctx, node)
	return config.String()
}

// GetSetting 获取设置
func (*sUtil) GetSetting(name string) (setting string, err error) {
	cacheKey := PublicCachePreFix + ":system_setting:" + name
	cached, err := service.Cache().GetCacheInstance().Get(Ctx, cacheKey)
	if err != nil {
		return "", err
	}
	//存在缓存key
	if !cached.IsNil() {
		return cached.String(), nil
	}
	//不存在缓存key，从数据库读取
	val, _ := g.Model("system_setting").Where("name", name).Value("value")
	err = service.Cache().GetCacheInstance().Set(Ctx, cacheKey, val.String(), 0)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

// ClearPublicCache 清除公共缓存
func (*sUtil) ClearPublicCache() (err error) {
	cachePreFix := PublicCachePreFix + ":"
	keys, err := service.Cache().GetCacheInstance().Keys(Ctx)
	if err != nil {
		return err
	}
	//g.Dump("keys", keys)
	for _, key := range keys {
		keyStr := gvar.New(key).String()
		if gstr.HasPrefix(keyStr, cachePreFix) {
			_, err = service.Cache().GetCacheInstance().Remove(Ctx, keyStr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ClearSystemSettingCache 清除后台设置缓存
func (*sUtil) ClearSystemSettingCache() (err error) {
	cachePreFix := PublicCachePreFix + ":system_setting:"
	keys, err := service.Cache().GetCacheInstance().Keys(Ctx)
	if err != nil {
		return err
	}
	for _, key := range keys {
		keyStr := gvar.New(key).String()
		if gstr.HasPrefix(keyStr, cachePreFix) {
			_, err = service.Cache().GetCacheInstance().Remove(Ctx, keyStr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetLocalIP 获取ip
func (*sUtil) GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String(), nil
	}
	return
}

// FriendyTimeFormat 计算时间差，并以"XXd XXh XXm XXs"返回
func (*sUtil) FriendyTimeFormat(TimeCreate time.Time, TimeEnd time.Time) string {
	SubTime := int(TimeEnd.Sub(TimeCreate).Seconds())
	// 秒
	if SubTime < 60 {
		return fmt.Sprintf("%d秒", SubTime)
	}
	// 分钟
	if SubTime < 60*60 {
		minute := int(math.Floor(float64(SubTime / 60)))
		second := SubTime % 60
		return fmt.Sprintf("%d分%d秒", minute, second)
	}
	// 小时
	if SubTime < 60*60*24 {
		hour := int(math.Floor(float64(SubTime / (60 * 60))))
		tail := SubTime % (60 * 60)
		minute := int(math.Floor(float64(tail / 60)))
		second := tail % 60
		return fmt.Sprintf("%d小时%d分%d秒", hour, minute, second)
	}
	// 天
	day := int(math.Floor(float64(SubTime / (60 * 60 * 24))))
	tail := SubTime % (60 * 60 * 24)
	hour := int(math.Floor(float64(tail / (60 * 60))))
	tail = SubTime % (60 * 60)
	minute := int(math.Floor(float64(tail / 60)))
	second := tail % 60
	return fmt.Sprintf("%d天%d小时%d分%d秒", day, hour, minute, second)
}

// ImageOrDefaultUrl 返回图片或默认图url
func (*sUtil) ImageOrDefaultUrl(imgUrl string) string {
	if imgUrl == "" {
		return "/resource/images/no_pic.jpg"
	}
	return imgUrl
}

// IsMobile 判断是手机端
func (s *sUtil) IsMobile(ctx context.Context) bool {
	userAgent := g.RequestFromCtx(ctx).UserAgent()
	if len(userAgent) == 0 {
		return false
	}
	isMobile := false
	mobileKeywords := []string{"Mobile", "Android", "Silk/", "Kindle",
		"BlackBerry", "Opera Mini", "Opera Mobi"}
	for i := 0; i < len(mobileKeywords); i++ {
		if strings.Contains(userAgent, mobileKeywords[i]) {
			isMobile = true
			break
		}
	}
	return isMobile
}

// ResponsiveJump 响应跳转
func (s *sUtil) ResponsiveJump(ctx context.Context) {
	// 获取配置的域名
	pcHost := service.Util().GetConfig("server.pcHost")
	mobileHost := service.Util().GetConfig("server.mobileHost")
	if mobileHost == "" {
		return
	}
	host := g.RequestFromCtx(ctx).GetHost()
	uri := g.RequestFromCtx(ctx).RequestURI
	fullUrl := g.RequestFromCtx(ctx).GetUrl()
	jumpUrl := ""
	if service.Util().IsMobile(ctx) {
		// 是手机访问
		if host != mobileHost {
			if gstr.Contains(fullUrl, mobileHost) {
				return
			}
			// 当前访问的域名不是手机域名，跳转手机域名对应路由
			jumpUrl = mobileHost + uri
		}
	} else {
		// 是pc访问
		if host == mobileHost {
			// 当前访问的域名是手机域名，跳转pc域名对应路由
			jumpUrl = pcHost + uri
		} else if gstr.Contains(fullUrl, mobileHost) {
			jumpUrl = gstr.Replace(fullUrl, mobileHost, pcHost, 1)
		}
	}
	if len(jumpUrl) > 0 {
		if !gstr.HasPrefix(jumpUrl, "http") {
			jumpUrl = "http://" + jumpUrl
		}
		g.RequestFromCtx(ctx).Response.RedirectTo(jumpUrl)
	}
	return
}
