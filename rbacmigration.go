package main

import (
	"github.com/dhanarJkusuma/pager"
	"net/http"
)

type AdminRoleMigration struct {
	auth *pager.Auth
}

func (a *AdminRoleMigration) Run(ptx *pager.PagerTx) error {
	var err error
	// create role
	adminRole := &pager.Role{
		Name:        "Administrator",
		Description: "Manage quiz data",
	}
	err = adminRole.CreateRole()
	if err != nil {
		return err
	}

	// set permission
	dashboardPage := &pager.Permission{
		Name:        "Dashboard Page",
		Method:      http.MethodGet,
		Route:       "/admin/quiz",
		Description: "Question Page Management",
	}
	dashboardPage.CreatePermission()
	adminRole.AddChild(dashboardPage)

	addQuestion := &pager.Permission{
		Name:        "Add Question",
		Method:      http.MethodPost,
		Route:       "/api/admin/quiz",
		Description: "Insert question for Admin API",
	}
	addQuestion.CreatePermission()
	adminRole.AddChild(addQuestion)

	getQuestion := &pager.Permission{
		Name:        "Get Question",
		Method:      http.MethodGet,
		Route:       "/api/admin/quiz",
		Description: "Get question for admin API",
	}
	getQuestion.CreatePermission()
	adminRole.AddChild(getQuestion)

	fetchDetail := &pager.Permission{
		Name:        "Fetch Detail Question",
		Method:      http.MethodGet,
		Route:       "/api/admin/quiz/detail",
		Description: "Insert question for admin API",
	}
	fetchDetail.CreatePermission()
	adminRole.AddChild(fetchDetail)

	toggleStatus := &pager.Permission{
		Name:        "Update Question Status",
		Method:      http.MethodPut,
		Route:       "/api/admin/quiz/status",
		Description: "Update status question for admin API",
	}
	toggleStatus.CreatePermission()
	adminRole.AddChild(toggleStatus)

	updateQuestion := &pager.Permission{
		Name:        "Update Question",
		Method:      http.MethodPut,
		Route:       "/api/admin/quiz/update",
		Description: "Update question for admin API",
	}
	updateQuestion.CreatePermission()
	adminRole.AddChild(updateQuestion)

	deleteQuestion := &pager.Permission{
		Name:        "Delete Question",
		Method:      http.MethodDelete,
		Route:       "/api/admin/quiz/delete",
		Description: "Delete question for admin API",
	}
	deleteQuestion.CreatePermission()
	adminRole.AddChild(deleteQuestion)

	answerUpdate := &pager.Permission{
		Name:        "Update Single Answer",
		Method:      http.MethodPut,
		Route:       "/api/admin/answer/update",
		Description: "Update single answer for admin API",
	}
	answerUpdate.CreatePermission()
	adminRole.AddChild(answerUpdate)

	deleteAnswer := &pager.Permission{
		Name:        "Delete Single Answer",
		Method:      http.MethodDelete,
		Route:       "/api/admin/answer/delete",
		Description: "Delete single answer for admin API",
	}
	deleteAnswer.CreatePermission()
	adminRole.AddChild(deleteAnswer)

	adminUser := &pager.User{
		Username: "administrator",
		Email:    "administrator@quizplatform.com",
		Password: "superadmin",
	}
	err = a.auth.Register(adminUser)
	if err != nil {
		return err
	}

	err = adminRole.Assign(adminUser)
	if err != nil {
		return err
	}
	return nil
}
