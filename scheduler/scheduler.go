package scheduler

import (
	"jwt_refresher/database"
	"jwt_refresher/models"
	"jwt_refresher/refresher"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	db     *database.DB
	engine *refresher.Engine
	ticker *time.Ticker
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewScheduler(db *database.DB, engine *refresher.Engine) *Scheduler {
	return &Scheduler{
		db:     db,
		engine: engine,
		stopCh: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")
	s.ticker = time.NewTicker(1 * time.Minute)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// 启动时立即执行一次检查
		s.checkAndRefresh()

		for {
			select {
			case <-s.ticker.C:
				s.checkAndRefresh()
			case <-s.stopCh:
				log.Println("Scheduler stopped")
				return
			}
		}
	}()

	log.Println("Scheduler started")
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
	s.wg.Wait()
	log.Println("Scheduler stopped successfully")
}

func (s *Scheduler) checkAndRefresh() {
	projects, err := s.db.GetEnabledProjects()
	if err != nil {
		log.Printf("Error getting enabled projects: %v", err)
		return
	}

	if len(projects) == 0 {
		return
	}

	log.Printf("Checking %d enabled projects for refresh...", len(projects))

	for _, project := range projects {
		if s.engine.ShouldRefresh(project) {
			log.Printf("Project %s (ID: %d) needs refresh", project.Name, project.ID)
			// 在goroutine中执行刷新，避免阻塞
			go func(p *models.Project) {
				if err := s.engine.Refresh(p); err != nil {
					log.Printf("Error refreshing project %s (ID: %d): %v", p.Name, p.ID, err)
				}
			}(project)
		}
	}
}
