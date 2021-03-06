package dept

import (
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"strings"
)

// Fill with you ideas below.

// Entity is the golang structure for table sys_dept.
type EntityExtend struct {
	DeptId     int64       `orm:"dept_id,primary" json:"dept_id"`     // 部门id
	ParentId   int64       `orm:"parent_id"       json:"parent_id"`   // 父部门id
	Ancestors  string      `orm:"ancestors"       json:"ancestors"`   // 祖级列表
	DeptName   string      `orm:"dept_name"       json:"dept_name"`   // 部门名称
	OrderNum   int         `orm:"order_num"       json:"order_num"`   // 显示顺序
	Leader     string      `orm:"leader"          json:"leader"`      // 负责人
	Phone      string      `orm:"phone"           json:"phone"`       // 联系电话
	Email      string      `orm:"email"           json:"email"`       // 邮箱
	Status     string      `orm:"status"          json:"status"`      // 部门状态（0正常 1停用）
	DelFlag    string      `orm:"del_flag"        json:"del_flag"`    // 删除标志（0代表存在 2代表删除）
	CreateBy   string      `orm:"create_by"       json:"create_by"`   // 创建者
	CreateTime *gtime.Time `orm:"create_time"     json:"create_time"` // 创建时间
	UpdateBy   string      `orm:"update_by"       json:"update_by"`   // 更新者
	UpdateTime *gtime.Time `orm:"update_time"     json:"update_time"` // 更新时间
	ParentName string      `json:"parentName"`                        // 父菜单名称
}

//分页请求参数
type SelectPageReq struct {
	ParentId  int64  `p:"parentId"`      //父部门ID
	DeptName  string `p:"deptName"`      //部门名称
	Status    string `p:"status"`        //状态
	BeginTime string `p:"beginTime"`     //开始时间
	EndTime   string `p:"endTime"`       //结束时间
	PageNum   int    `p:"pageNum"`       //当前页码
	PageSize  int    `p:"pageSize"`      //每页数
	SortName  string `p:"orderByColumn"` //排序字段
	SortOrder string `p:"isAsc"`         //排序方式
}

//新增页面请求参数
type AddReq struct {
	ParentId int64  `p:"parentId"  v:"required#父节点不能为空"`
	DeptName string `p:"deptName"  v:"required#部门名称不能为空"`
	OrderNum int    `p:"orderNum" v:"required#显示排序不能为空"`
	Leader   string `p:"leader"`
	Phone    string `p:"phone"`
	Email    string `p:"email"`
	Status   string `p:"status"`
}

//修改页面请求参数
type EditReq struct {
	DeptId   int64  `p:"deptId" v:"required#主键ID不能为空"`
	ParentId int64  `p:"parentId"  v:"required#父节点不能为空"`
	DeptName string `p:"deptName"  v:"required#部门名称不能为空"`
	OrderNum int    `p:"orderNum" v:"required#显示排序不能为空"`
	Leader   string `p:"leader"`
	Phone    string `p:"phone"`
	Email    string `p:"email"`
	Status   string `p:"status"`
}

//检查菜单名称请求参数
type CheckDeptNameReq struct {
	DeptId   int64  `p:"deptId"  v:"required#部门ID不能为空"`
	ParentId int64  `p:"parentId"  v:"required#父部门ID不能为空"`
	DeptName string `p:"deptName"  v:"required#部门名称不能为空"`
}

//检查菜单名称请求参数
type CheckDeptNameALLReq struct {
	ParentId int64  `p:"parentId"  v:"required#父部门ID不能为空"`
	DeptName string `p:"deptName"  v:"required#部门名称不能为空"`
}

//根据部门ID查询信息
func SelectDeptById(id int64) (*EntityExtend, error) {
	db, err := gdb.Instance()

	if err != nil {
		return nil, gerror.New("获取数据库连接失败")
	}
	var result EntityExtend
	model := db.Table("sys_dept d")
	model.Fields("d.dept_id, d.parent_id, d.ancestors, d.dept_name, d.order_num, d.leader, d.phone, d.email, d.status,(select dept_name from sys_dept where dept_id = d.parent_id) parent_name")
	model.Where("d.dept_id = ?", id)
	err = model.Struct(&result)
	return &result, err
}

//根据ID查询所有子部门
func SelectChildrenDeptById(deptId int64) []*Entity {
	rs, _ := FindAll("find_in_set(?, ancestors)", deptId)
	return rs
}

//删除部门管理信息
func DeleteDeptById(deptId int64) int64 {
	rs, err := Update("del_flag = '2'", "dept_id = ?", deptId)
	if err != nil {
		return 0
	}

	rows, _ := rs.RowsAffected()

	return rows

}

//修改子元素关系
func UpdateDeptChildren(deptId int64, newAncestors, oldAncestors string) {
	deptList := SelectChildrenDeptById(deptId)

	if deptList == nil || len(deptList) <= 0 {
		return
	}

	for _, tmp := range deptList {
		tmp.Ancestors = strings.ReplaceAll(tmp.Ancestors, oldAncestors, newAncestors)
	}

	ancestors := " case dept_id"
	idStr := ""

	for _, dept := range deptList {
		ancestors += " when " + gconv.String(dept.DeptId) + " then " + dept.Ancestors
		if idStr == "" {
			idStr = gconv.String(dept.DeptId)
		} else {
			idStr += "," + gconv.String(dept.DeptId)
		}
	}

	ancestors += " end "

	Update("ancestors = ?", "dept_id in(?)", ancestors, idStr)
}

//查询部门管理数据
func SelectDeptList(parentId int64, deptName, status string) (*[]Entity, error) {
	var result []Entity
	db, err := gdb.Instance()
	if err != nil {
		return nil, gerror.New("获取数据库连接失败")
	}
	model := db.Table("sys_dept d").Where("d.del_flag = '0'")
	if parentId > 0 {
		model.Where("d.parent_id = ?", parentId)
	}
	if deptName != "" {
		model.Where("d.dept_name like ?", "%"+deptName+"%")
	}
	if status != "" {
		model.Where("d.status = ?", status)
	}
	model.Order("d.parent_id, d.order_num")

	err = model.Structs(&result)

	return &result, err
}

//根据角色ID查询部门
func SelectRoleDeptTree(roleId int64) (*[]string, error) {
	db, err := gdb.Instance()
	if err != nil {
		return nil, gerror.New("获取数据库连接失败")
	}
	model := db.Table("sys_dept d").LeftJoin("sys_role_dept rd", "d.dept_id = rd.dept_id")
	model.Where("d.del_flag = '0'").Where("rd.role_id = ?", roleId)
	model.Order("d.parent_id, d.order_num")
	model.Fields("concat(d.dept_id, d.dept_name) as dept_name")
	var result []string
	rs, err := model.All()
	if err == nil && rs != nil && len(rs) > 0 {
		for _, record := range rs {
			if record["dept_name"].String() != "" {
				result = append(result, record["dept_name"].String())
			}
		}
	}
	return &result, nil
}

//查询部门是否存在用户
func CheckDeptExistUser(deptId int64) bool {
	num, _ := FindCount("dept_id = ? and del_flag = '0'", deptId)
	if num > 0 {
		return true
	} else {
		return false
	}
}

//查询部门人数
func SelectDeptCount(deptId, parentId int64) int {
	result := 0
	whereStr := "del_flag = '0'"
	if deptId > 0 {
		whereStr = whereStr + " and dept_id=" + gconv.String(deptId)
	}
	if parentId > 0 {
		whereStr = whereStr + " and parent_id=" + gconv.String(parentId)
	}
	rs, err := FindCount(whereStr)
	if err != nil {
		result = rs
	}
	return result
}

//校验部门名称是否唯一
func CheckDeptNameUnique(deptName string, deptId, parentId int64) (*Entity, error) {
	return FindOne("dept_id !=? and dept_name=? and parent_id = ?", deptId, deptName, parentId)
}

//校验部门名称是否唯一
func CheckDeptNameUniqueAll(deptName string, parentId int64) (*Entity, error) {
	return FindOne("dept_name=? and parent_id = ?", deptName, parentId)
}
