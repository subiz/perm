package perm

import (
	"bitbucket.org/subiz/header/auth"
	"context"
	"bitbucket.org/subiz/gocommon"
	"bitbucket.org/subiz/header/lang"
	"bitbucket.org/subiz/auth/scope"
)

type rule struct {
	agentid string
	method *auth.Method
}

type Checker struct {
	rules []rule
	db DB
}

func (c *Checker) Or(agentid string, method *auth.Method) *Checker {
	c.rules = append(c.rules, rule{
		agentid: agentid,
		method: method,
	})
	return c
}

func (me Perm) New() *Checker {
	return &Checker{
		rules: make([]rule, 0),
		db: me.db,
	}
}

func (c *Checker) Check(ctx context.Context, accid string) {
	cred := common.GetCredential(ctx)
	if cred.GetAccountId() == "" {
		cred = nil
	}
	c.CheckCred(cred, accid)
}

func (c *Checker) CheckCred(cred *auth.Credential, accid string) {
	if cred == nil {
		panic(common.New400(lang.T_invalid_credential))
	}

	if accid != "" && cred.GetAccountId() != accid {
		panic(common.New400(lang.T_wrong_account_in_credential))
	}

	issuer := cred.GetIssuer()
	if issuer == "" {
		panic(common.New400(lang.T_invalid_credential))
	}

	usermethod := c.db.Read(cred.GetAccountId(), issuer)
	clientmethod := *cred.GetMethod()
	realmethod := scope.IntersectMethod(clientmethod, usermethod)

	for _, r := range c.rules {
		if r.agentid != "" && r.agentid != issuer {
			panic(common.New400(lang.T_wrong_user_in_credential))
		}

		if scope.RequireMethod(realmethod, *r.method) {
			return
		}
	}
	panic(common.New400(lang.T_access_deny))
}