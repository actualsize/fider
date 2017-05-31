package postgres_test

import (
	"testing"
	"time"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/dbx"
	"github.com/getfider/fider/app/storage/postgres"
	. "github.com/onsi/gomega"
)

func TestIdeaStorage_GetAll(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	now := time.Now()

	trx.Execute("INSERT INTO ideas (title, slug, number, description, created_on, tenant_id, user_id, supporters, status) VALUES ('Idea #1', 'idea-1', 1, 'Description #1', $1, 300, 300, 0, 1)", now)
	trx.Execute("INSERT INTO ideas (title, slug, number, description, created_on, tenant_id, user_id, supporters, status) VALUES ('Idea #2', 'idea-2', 2, 'Description #2', $1, 300, 301, 5, 2)", now)

	ideas := &postgres.IdeaStorage{Trx: trx}
	dbIdeas, err := ideas.GetAll(300)

	Expect(err).To(BeNil())
	Expect(dbIdeas).To(HaveLen(2))

	Expect(dbIdeas[0].Title).To(Equal("Idea #1"))
	Expect(dbIdeas[0].Slug).To(Equal("idea-1"))
	Expect(dbIdeas[0].Number).To(Equal(1))
	Expect(dbIdeas[0].Description).To(Equal("Description #1"))
	Expect(dbIdeas[0].User.Name).To(Equal("Jon Snow"))
	Expect(dbIdeas[0].TotalSupporters).To(Equal(0))
	Expect(dbIdeas[0].Status).To(Equal(models.IdeaStarted))

	Expect(dbIdeas[1].Title).To(Equal("Idea #2"))
	Expect(dbIdeas[1].Slug).To(Equal("idea-2"))
	Expect(dbIdeas[1].Number).To(Equal(2))
	Expect(dbIdeas[1].Description).To(Equal("Description #2"))
	Expect(dbIdeas[1].User.Name).To(Equal("Arya Stark"))
	Expect(dbIdeas[1].TotalSupporters).To(Equal(5))
	Expect(dbIdeas[1].Status).To(Equal(models.IdeaCompleted))
}

func TestIdeaStorage_SaveAndGet(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	trx.Execute("INSERT INTO tenants (name, subdomain, created_on) VALUES ('My Domain Inc.','mydomain', now())")
	trx.Execute("INSERT INTO users (name, email, created_on, role) VALUES ('Jon Snow','jon.snow@got.com', now(), 2)")

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, err := ideas.Save(1, 1, "My new idea", "with this description")
	Expect(err).To(BeNil())
	Expect(idea.ID).To(Equal(1))

	dbIdea, err := ideas.GetByID(1, 1)

	Expect(err).To(BeNil())
	Expect(dbIdea.ID).To(Equal(1))
	Expect(dbIdea.Number).To(Equal(1))
	Expect(dbIdea.TotalSupporters).To(Equal(0))
	Expect(dbIdea.Status).To(Equal(models.IdeaNew))
	Expect(dbIdea.Title).To(Equal("My new idea"))
	Expect(dbIdea.Description).To(Equal("with this description"))
	Expect(dbIdea.User.ID).To(Equal(1))
	Expect(dbIdea.User.Name).To(Equal("Jon Snow"))
	Expect(dbIdea.User.Email).To(Equal("jon.snow@got.com"))
}

func TestIdeaStorage_GetInvalid(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	dbIdea, err := ideas.GetByID(1, 1)

	Expect(err).To(Equal(app.ErrNotFound))
	Expect(dbIdea).To(BeNil())
}

func TestIdeaStorage_AddAndReturnComments(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	trx.Execute("INSERT INTO tenants (name, subdomain, created_on) VALUES ('My Domain Inc.','mydomain', now())")
	trx.Execute("INSERT INTO users (name, email, created_on, role) VALUES ('Jon Snow','jon.snow@got.com', now(), 2)")

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(1, 1, "My new idea", "with this description")
	ideas.AddComment(1, idea.ID, "Comment #1")
	ideas.AddComment(1, idea.ID, "Comment #2")

	comments, err := ideas.GetCommentsByIdeaID(1, idea.ID)
	Expect(err).To(BeNil())
	Expect(len(comments)).To(Equal(2))

	Expect(comments[0].Content).To(Equal("Comment #2"))
	Expect(comments[1].Content).To(Equal("Comment #1"))
}

func TestIdeaStorage_SaveAndGet_DifferentTenants(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	ideas.Save(300, 300, "My new idea", "with this description")
	ideas.Save(400, 400, "My other idea", "with other description")

	dbIdea, err := ideas.GetByNumber(300, 1)

	Expect(err).To(BeNil())
	Expect(dbIdea.ID).To(Equal(1))
	Expect(dbIdea.Number).To(Equal(1))
	Expect(dbIdea.Title).To(Equal("My new idea"))
	Expect(dbIdea.Slug).To(Equal("my-new-idea"))

	dbIdea, err = ideas.GetByNumber(400, 1)

	Expect(err).To(BeNil())
	Expect(dbIdea.ID).To(Equal(2))
	Expect(dbIdea.Number).To(Equal(1))
	Expect(dbIdea.Title).To(Equal("My other idea"))
	Expect(dbIdea.Slug).To(Equal("my-other-idea"))

}

func TestIdeaStorage_AddSupporter(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")

	err := ideas.AddSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	err = ideas.AddSupporter(300, 301, idea.ID)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(2))
}

func TestIdeaStorage_AddSupporter_Twice(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")

	err := ideas.AddSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	err = ideas.AddSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(1))
}

func TestIdeaStorage_RemoveSupporter(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")

	err := ideas.AddSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_RemoveSupporter_Twice(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")

	err := ideas.AddSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(300, 300, idea.ID)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_SetResponse(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")
	err := ideas.SetResponse(300, idea.ID, "We liked this idea", 300, models.IdeaStarted)

	Expect(err).To(BeNil())

	idea, _ = ideas.GetByID(300, idea.ID)
	Expect(idea.Response.Text).To(Equal("We liked this idea"))
	Expect(idea.Status).To(Equal(models.IdeaStarted))
	Expect(idea.Response.User.ID).To(Equal(300))
}

func TestIdeaStorage_AddSupporter_ClosedIdea(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")
	ideas.SetResponse(300, idea.ID, "We liked this idea", 300, models.IdeaCompleted)
	ideas.AddSupporter(300, 300, idea.ID)

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_RemoveSupporter_ClosedIdea(t *testing.T) {
	RegisterTestingT(t)
	db, _ := dbx.New()
	db.Seed()
	defer db.Close()

	trx, _ := db.Begin()
	defer trx.Rollback()

	ideas := &postgres.IdeaStorage{Trx: trx}
	idea, _ := ideas.Save(300, 300, "My new idea", "with this description")
	ideas.AddSupporter(300, 300, idea.ID)
	ideas.SetResponse(300, idea.ID, "We liked this idea", 300, models.IdeaCompleted)
	ideas.RemoveSupporter(300, 300, idea.ID)

	dbIdea, err := ideas.GetByNumber(300, 1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(1))
}
