package permission

import (
	"context"
	"gf_cms/internal/consts"
	"gf_cms/internal/dao"
	"gf_cms/internal/logic/admin"
	"gf_cms/internal/logic/casbinPolicy"
	"gf_cms/internal/logic/util"
	"gf_cms/internal/model"
	"gf_cms/internal/model/entity"
	"gf_cms/internal/service"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type sPermission struct{}

var (
	insPermission = sPermission{}
)

func init() {
	service.RegisterPermission(New())
}

func New() *sPermission {
	return &sPermission{}
}

// Permission 权限
func Permission() *sPermission {
	return &insPermission
}

func (*sPermission) readYaml(ctx context.Context) (conf *model.PermissionConfig, err error) {
	data, err := g.Cfg("permission").Data(ctx)
	if err != nil {
		return nil, err
	}
	err = gconv.Scan(data, &conf)
	if err != nil {
		return nil, err
	}
	return
}

// BackendAll 获取后台全部权限（view和api）
func (*sPermission) BackendAll() (res []model.PermissionAllItem, err error) {
	backendViewAll, err := service.Permission().BackendViewAll()
	if err != nil {
		return nil, err
	}
	backendApiAll, err := service.Permission().BackendApiAll()
	if err != nil {
		return nil, err
	}
	var permissionAll []model.PermissionAllItem
	for _, viewItem := range backendViewAll {
		// 只有视图的slug在permissionAll中不存在，直接将只有视图的权限放进去
		if Permission().slugInModelPermissionAllItem(viewItem.Slug, permissionAll) == false {
			permissionAllItem := model.PermissionAllItem{}
			permissionAllItem.Title = viewItem.Title
			permissionAllItem.Slug = viewItem.Slug
			permissionAllItem.BackendViewPermissions = viewItem.Permissions
			permissionAll = append(permissionAll, permissionAllItem)
		} else {
			for key, permissionAllItem := range permissionAll {
				if permissionAllItem.Slug == viewItem.Slug {
					permissionAll[key].BackendViewPermissions = viewItem.Permissions
				}
			}
		}
	}
	for _, apiItem := range backendApiAll {
		// 只有接口的slug在permissionAll中不存在，直接将只有接口的权限放进去
		if Permission().slugInModelPermissionAllItem(apiItem.Slug, permissionAll) == false {
			permissionAllItem := model.PermissionAllItem{}
			permissionAllItem.Title = apiItem.Title
			permissionAllItem.Slug = apiItem.Slug
			permissionAllItem.BackendApiPermissions = apiItem.Permissions
			permissionAll = append(permissionAll, permissionAllItem)
		} else {
			for key, permissionAllItem := range permissionAll {
				if permissionAllItem.Slug == apiItem.Slug {
					permissionAll[key].BackendApiPermissions = apiItem.Permissions
				}
			}
		}
	}
	return permissionAll, nil
}

// BackendViewAll Backend 获取全部后台权限
func (*sPermission) BackendViewAll() (res []model.PermissionGroups, err error) {
	cacheKey := util.PublicCachePreFix + ":permissions:backend_view_all"
	result, err := g.Redis().Do(util.Ctx, "GET", cacheKey)
	if err != nil {
		return nil, err
	}
	if !result.IsEmpty() {
		var permissionGroups []model.PermissionGroups
		if err = result.Structs(&permissionGroups); err != nil {
			return nil, err
		}
		return permissionGroups, nil
	}
	conf, _ := Permission().readYaml(util.Ctx)
	backendViewAll := conf.BackendView.Groups
	_, err = g.Redis().Do(util.Ctx, "SET", cacheKey, backendViewAll)
	if err != nil {
		return nil, err
	}
	return backendViewAll, nil
}

// BackendApiAll Backend 获取全部后台接口权限
func (*sPermission) BackendApiAll() (res []model.PermissionGroups, err error) {
	cacheKey := util.PublicCachePreFix + ":permissions:backend_api_all"
	result, err := g.Redis().Do(util.Ctx, "GET", cacheKey)
	if err != nil {
		return nil, err
	}
	if !result.IsEmpty() {
		var permissionGroups []model.PermissionGroups
		if err = result.Structs(&permissionGroups); err != nil {
			return nil, err
		}
		return permissionGroups, nil
	}
	conf, _ := Permission().readYaml(util.Ctx)
	backendApiAll := conf.BackendApi.Groups
	_, err = g.Redis().Do(util.Ctx, "SET", cacheKey, backendApiAll)
	if err != nil {
		return nil, err
	}
	return backendApiAll, nil
}

// BackendMyView 获取我的所有后台视图权限
func (*sPermission) BackendMyView(accountId string) (myPermissions []gdb.Value, err error) {
	roleIds, err := admin.Admin().GetRoleIdsByAccountId(accountId)
	if len(roleIds) == 0 {
		return nil, gerror.New("用户无任何角色")
	}
	myPermissions, err = dao.CmsRulePermissions.Ctx(util.Ctx).
		Where("p_type", "p").
		WhereIn("v0", roleIds).
		Where("v1", "backend").
		Fields("v2").
		Array()
	if err != nil {
		return nil, err
	}
	return myPermissions, nil
}

// BackendMyApi 获取我的所有后台接口权限
func (*sPermission) BackendMyApi(accountId string) (myPermissions []gdb.Value, err error) {
	roleIds, err := admin.Admin().GetRoleIdsByAccountId(accountId)
	if len(roleIds) == 0 {
		return nil, gerror.New("用户无任何角色")
	}
	myPermissions, err = dao.CmsRulePermissions.Ctx(util.Ctx).
		Where("p_type", "p").
		WhereIn("v0", roleIds).
		Where("v1", "backend_api").
		Fields("v2").
		Array()
	if err != nil {
		return nil, err
	}
	return myPermissions, nil
}

// GetAllViewPermissionsArray 获取全部视图权限数组
func (*sPermission) GetAllViewPermissionsArray() (res []string, err error) {
	backendViewAllPermissions, err := service.Permission().BackendViewAll()
	if err != nil {
		return nil, err
	}
	var permissionsArray = make([]string, 0, len(backendViewAllPermissions))
	for _, _item := range backendViewAllPermissions {
		for _, _permission := range _item.Permissions {
			permission := gconv.String(_item.Slug) + "." + gconv.String(_permission.Slug)
			permissionsArray = append(permissionsArray, permission)
		}
	}
	return permissionsArray, nil
}

// GetAllApiPermissionsArray 获取全部接口权限数组
func (*sPermission) GetAllApiPermissionsArray() (res []string, err error) {
	backendApiAllPermissions, err := service.Permission().BackendApiAll()
	if err != nil {
		return nil, err
	}
	var permissionsArray = make([]string, 0, len(backendApiAllPermissions))
	for _, _item := range backendApiAllPermissions {
		for _, _permission := range _item.Permissions {
			permission := gconv.String(_item.Slug) + "." + gconv.String(_permission.Slug)
			permissionsArray = append(permissionsArray, permission)
		}
	}
	return permissionsArray, nil
}

// BackendUserViewCan 检测后台用户有无视图的操作权限
func (*sPermission) BackendUserViewCan(ctx context.Context, routePermission string) bool {
	cmsAdmin, err := Permission().BackendGetUserFromSession(ctx)
	if err != nil {
		return false
	}
	obj := casbinPolicy.CasbinPolicy().ObjBackend()
	return casbinPolicy.CasbinPolicy().CheckByAccountId(gvar.New(cmsAdmin.Id).String(), obj, routePermission)
}

// BackendUserApiCan 检测后台用户有无接口的操作权限
func (*sPermission) BackendUserApiCan(ctx context.Context, routePermission string) bool {
	cmsAdmin, err := Permission().BackendGetUserFromSession(ctx)
	if err != nil {
		return false
	}
	obj := casbinPolicy.CasbinPolicy().ObjBackendApi()
	return casbinPolicy.CasbinPolicy().CheckByAccountId(gvar.New(cmsAdmin.Id).String(), obj, routePermission)
}

// BackendGetUserFromSession 从session获取当前登录用户
func (s *sPermission) BackendGetUserFromSession(ctx context.Context) (out *entity.CmsAdmin, err error) {
	var adminSession, _ = g.RequestFromCtx(ctx).Session.Get(consts.AdminSessionKeyPrefix)
	var cmsAdmin *entity.CmsAdmin
	err = gconv.Scan(adminSession, &cmsAdmin)
	if err != nil {
		return nil, err
	}
	return cmsAdmin, nil
}

// 判断slug是否在model.PermissionGroups数组中
func (*sPermission) slugInModelPermissionGroups(slug string, permissionGroups []model.PermissionGroups) bool {
	for _, item := range permissionGroups {
		if slug == item.Slug {
			return true
		}
	}
	return false
}

// 判断slug是否在model.PermissionAllItem数组中
func (*sPermission) slugInModelPermissionAllItem(slug string, permissionGroups []model.PermissionAllItem) bool {
	for _, item := range permissionGroups {
		if slug == item.Slug {
			return true
		}
	}
	return false
}
