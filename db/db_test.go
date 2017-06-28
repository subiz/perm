package db_test

// NOTE: set up cassandra first:
// docker run --name chat-cassandra -d -p 9042:9042 cassandra:3.10

import (
	"testing"
	. "bitbucket.org/subiz/perm/db"
	pb "bitbucket.org/subiz/servicespec/proto/auth"
	"time"
	"fmt"
	"strings"
	"strconv"
	scope "bitbucket.org/subiz/scopemgr"
)

func skipPermTest(t *testing.T) {
	//t.Skip()
}

var db *PermDB
func tearUpPermTest(t *testing.T) {
	db = NewPermDB([]string{"127.0.0.1:9042"}, "testperm", 1)
}

func TestPermCrud(t *testing.T) {
	skipPermTest(t)
	tearUpPermTest(t)

	testmethod := &pb.Method{
		ReadAccount: true,
		ReadAgents: true,
	}
	db.Update("acc_permcrud", "user_permcrud", testmethod)
	method := db.Read("acc_permcrud", "user_permcrud")
	if !scope.RequireMethod(method, testmethod) || !scope.RequireMethod(testmethod, method) {
		t.Fatal("should be equal")
	}
}

func TestListPerms(t *testing.T) {
	skipPermTest(t)
	tearUpPermTest(t)
	var now = time.Now().UnixNano()
	// add 3000 method in same account
	N, Ndisabled, Nwrong := 3000, 1500, 800
	accid := fmt.Sprintf("acc_listperm_%d", now)
	for i := 0; i < N; i++ {
		if i >= N - Nwrong {
			db.Update(accid, fmt.Sprintf("user_listperm_%d", i), &pb.Method{})
			continue
		}
		db.Update(accid, fmt.Sprintf("user_listperm_%d", i), &pb.Method{
			ReadAgents: true,
		})
	}
	// disable 1500 account
	for i := 0; i < Ndisabled; i++ {
		db.UpdateState(accid, fmt.Sprintf("user_listperm_%d", i), false)
	}
	ids := db.ListUsersByMethod(accid, &pb.Method{
		ReadAgents: true,
	}, "", N)

	if len(ids) != N - Ndisabled - Nwrong {
		t.Fatalf("missing user, got %d", len(ids))
	}
	for _, id := range ids {
		i := strings.Split(id, "_")[2]
		if i < strconv.Itoa(Ndisabled) {
			t.Fatal("should not include inactive agent")
		}
		if i >= strconv.Itoa(N - Nwrong) {
			t.Fatalf("should not include wrong agent, got id %s", id)
		}
	}
}
