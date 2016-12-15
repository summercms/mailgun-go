package mailgun

import (
	"fmt"
	"strings"
	"testing"
)

func setup(t *testing.T) (Mailgun, string) {
	domain := reqEnv(t, "MG_DOMAIN")
	mg, err := NewMailgunFromEnv()
	if err != nil {
		t.Fatalf("NewMailgunFromEnv() error - %s", err.Error())
	}

	address := fmt.Sprintf("%s@%s", strings.ToLower(randomString(6, "list")), domain)
	_, err = mg.CreateList(List{
		Address:     address,
		Name:        address,
		Description: "TestMailingListMembers-related mailing list",
		AccessLevel: Members,
	})
	if err != nil {
		t.Fatal(err)
	}
	return mg, address
}

func teardown(t *testing.T, mg Mailgun, address string) {
	err := mg.DeleteList(address)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMailingListMembers(t *testing.T) {
	mg, address := setup(t)
	defer teardown(t, mg, address)

	var countPeople = func() int {
		n, _, err := mg.GetMembers(DefaultLimit, DefaultSkip, All, address)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}

	startCount := countPeople()
	protoJoe := Member{
		Address:    "joe@example.com",
		Name:       "Joe Example",
		Subscribed: Subscribed,
	}
	err := mg.CreateMember(true, address, protoJoe)
	if err != nil {
		t.Fatal(err)
	}

	newCount := countPeople()
	if newCount <= startCount {
		t.Fatalf("Expected %d people subscribed; got %d", startCount+1, newCount)
	}

	theMember, err := mg.GetMemberByAddress("joe@example.com", address)
	if err != nil {
		t.Fatal(err)
	}
	if (theMember.Address != protoJoe.Address) ||
		(theMember.Name != protoJoe.Name) ||
		(*theMember.Subscribed != *protoJoe.Subscribed) ||
		(len(theMember.Vars) != 0) {
		t.Fatalf("Unexpected Member: Expected [%#v], Got [%#v]", protoJoe, theMember)
	}

	_, err = mg.UpdateMember("joe@example.com", address, Member{
		Name: "Joe Cool",
	})
	if err != nil {
		t.Fatal(err)
	}

	theMember, err = mg.GetMemberByAddress("joe@example.com", address)
	if err != nil {
		t.Fatal(err)
	}
	if theMember.Name != "Joe Cool" {
		t.Fatal("Expected Joe Cool; got " + theMember.Name)
	}

	err = mg.DeleteMember("joe@example.com", address)
	if err != nil {
		t.Fatal(err)
	}

	if countPeople() != startCount {
		t.Fatalf("Expected %d people; got %d instead", startCount, countPeople())
	}

	err = mg.CreateMemberList(nil, address, []interface{}{
		Member{
			Address:    "joe.user1@example.com",
			Name:       "Joe's debugging account",
			Subscribed: Unsubscribed,
		},
		Member{
			Address:    "Joe Cool <joe.user2@example.com>",
			Name:       "Joe's Cool Account",
			Subscribed: Subscribed,
		},
		Member{
			Address: "joe.user3@example.com",
			Vars: map[string]interface{}{
				"packet-email": "KW9ABC @ BOGBBS-4.#NCA.CA.USA.NOAM",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	theMember, err = mg.GetMemberByAddress("joe.user2@example.com", address)
	if err != nil {
		t.Fatal(err)
	}
	if theMember.Name != "Joe's Cool Account" {
		t.Fatalf("Expected Joe's Cool Account; got %s", theMember.Name)
	}
	if theMember.Subscribed != nil {
		if *theMember.Subscribed != true {
			t.Fatalf("Expected subscribed to be true; got %v", *theMember.Subscribed)
		}
	} else {
		t.Fatal("Expected some kind of subscription status; got nil.")
	}
}

func TestMailingLists(t *testing.T) {
	domain := reqEnv(t, "MG_DOMAIN")
	mg, err := NewMailgunFromEnv()
	if err != nil {
		t.Fatalf("NewMailgunFromEnv() error - %s", err.Error())
	}
	listAddr := fmt.Sprintf("%s@%s", strings.ToLower(randomString(7, "list")), domain)
	protoList := List{
		Address:     listAddr,
		Name:        "List1",
		Description: "A list created by an acceptance test.",
		AccessLevel: Members,
	}

	var countLists = func() int {
		total, _, err := mg.GetLists(DefaultLimit, DefaultSkip, "")
		if err != nil {
			t.Fatal(err)
		}
		return total
	}

	_, err = mg.CreateList(protoList)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = mg.DeleteList(listAddr)
		if err != nil {
			t.Fatal(err)
		}

		theList, err := mg.GetListByAddress(listAddr)
		if err == nil {
			t.Fatalf("Expected list %s deleted", theList.Address)
		}
	}()

	actualCount := countLists()
	if actualCount < 1 {
		t.Fatalf("Expected atleast 1 lists defined; got %d", actualCount)
	}

	theList, err := mg.GetListByAddress(listAddr)
	if err != nil {
		t.Fatal(err)
	}
	protoList.CreatedAt = theList.CreatedAt // ignore this field when comparing.
	if theList != protoList {
		t.Fatalf("Unexpected list descriptor: Expected [%#v], Got [%#v]", protoList, theList)
	}

	_, err = mg.UpdateList(listAddr, List{
		Description: "A list whose description changed",
	})
	if err != nil {
		t.Fatal(err)
	}

	theList, err = mg.GetListByAddress(listAddr)
	if err != nil {
		t.Fatal(err)
	}
	newList := protoList
	newList.Description = "A list whose description changed"
	if theList != newList {
		t.Fatalf("Expected [%#v], Got [%#v]", newList, theList)
	}
}
