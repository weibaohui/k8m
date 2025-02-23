package user

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}

	// todo 管理页面，判断是否管理员，看到所有的用户，
	user, role := amis.GetLoginUser(c)
	var queryFuncs []func(*gorm.DB) *gorm.DB
	switch role {
	case models.RolePlatformAdmin:
		params.UserName = ""
		queryFuncs = []func(*gorm.DB) *gorm.DB{
			func(db *gorm.DB) *gorm.DB {
				return db
			},
		}
	case models.RoleClusterAdmin:
		queryFuncs = []func(*gorm.DB) *gorm.DB{
			func(db *gorm.DB) *gorm.DB {
				return db.Where("created_by=?", user)
			},
		}
	case models.RoleClusterReadonly:
		queryFuncs = []func(*gorm.DB) *gorm.DB{
			func(db *gorm.DB) *gorm.DB {
				return db.Where("username=?", user)
			},
		}
	}
	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 用户名不能为admin

	if m.Username == "admin" {
		amis.WriteJsonError(c, fmt.Errorf("用户名不能为admin"))
		return
	}

	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		if m.ID == 0 {
			// 新增
			return db
		} else {
			// 修改
			return db.Select([]string{"username", "role"}).Updates(m)
		}
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}
func Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.User{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func UpdatePsw(c *gin.Context) {
	// todo 校验管理员权限，不是管理员不可以改
	id := c.Param("id")
	params := dao.BuildParams(c)
	m := &models.User{}
	err := c.ShouldBindJSON(m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	m.ID = uint(utils.ToInt64(id))

	// 密码 + 盐重新计算
	pswBytes, err := utils.AesDecrypt(m.Password)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	m.Salt = utils.RandNLengthString(8)
	psw, err := utils.AesEncrypt([]byte(fmt.Sprintf("%s%s", string(pswBytes), m.Salt)))
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	m.Password = base64.StdEncoding.EncodeToString(psw)

	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"password", "salt"}).Updates(m)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
