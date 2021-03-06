package jiraui

import (
	"fmt"
	"regexp"

	ui "github.com/gizak/termui"
)

type TicketListPage struct {
	BaseListPage
	CommandBarFragment
	StatusBarFragment
	ActiveQuery Query
	ActiveSort  Sort
}

func (p *TicketListPage) Search() {
	s := p.ActiveSearch
	n := len(p.cachedResults)
	if s.command == "" {
		return
	}
	increment := 1
	if s.directionUp {
		increment = -1
	}
	// we use modulo here so we can loop through every line.
	// adding 'n' means we never have '-1 % n'.
	startLine := (p.uiList.Cursor + n + increment) % n
	for i := startLine; i != p.uiList.Cursor; i = (i + increment + n) % n {
		if s.re.MatchString(p.cachedResults[i]) {
			p.uiList.SetCursorLine(i)
			p.Update()
			break
		}
	}
}

func (p *TicketListPage) ActiveTicketId() string {
	return p.GetSelectedTicketId()
}

func (p *TicketListPage) GetSelectedTicketId() string {
	return findTicketIdInString(p.cachedResults[p.uiList.Cursor])
}

func (p *TicketListPage) SelectItem() {
	if len(p.cachedResults) == 0 {
		return
	}
	q := new(TicketShowPage)
	q.TicketId = p.GetSelectedTicketId()
	previousPages = append(previousPages, currentPage)
	currentPage = q
	q.Create()
	changePage()
}

func (p *TicketListPage) GoBack() {
	if len(previousPages) == 0 {
		currentPage = ticketQueryPage
	} else {
		currentPage, previousPages = previousPages[len(previousPages)-1], previousPages[:len(previousPages)-1]
	}
	changePage()
}

func (p *TicketListPage) EditTicket() {
	runJiraCmdEdit(p.GetSelectedTicketId())
}

func (p *TicketListPage) Update() {
	ui.Render(p.uiList)
	p.statusBar.Update()
	p.commandBar.Update()
}

func (p *TicketListPage) Refresh() {
	pDeref := &p
	q := *pDeref
	q.cachedResults = make([]string, 0)
	currentPage = q
	changePage()
	q.Create()
}

func (p *TicketListPage) Create() {
	log.Debugf("TicketListPage.Create(): self:        %s (%p)", p.Id(), p)
	log.Debugf("TicketListPage.Create(): currentPage: %s (%p)", currentPage.Id(), currentPage)
	ui.Clear()
	if p.uiList == nil {
		p.uiList = NewScrollableList()
	}
	if p.statusBar == nil {
		p.statusBar = new(StatusBar)
	}
	if p.commandBar == nil {
		p.commandBar = commandBar
	}
	query := p.ActiveQuery.JQL
	if sort := p.ActiveSort.JQL; sort != "" {
		re := regexp.MustCompile(`(?i)\s+ORDER\s+BY.+$`)
		query = re.ReplaceAllString(query, ``) + " " + sort
	}
	if len(p.cachedResults) == 0 {
		p.cachedResults = JiraQueryAsStrings(query, p.ActiveQuery.Template)
	}
	if p.uiList.Cursor >= len(p.cachedResults) {
		p.uiList.Cursor = len(p.cachedResults) - 1
	}
	p.uiList.Items = p.cachedResults
	p.uiList.ItemFgColor = ui.ColorYellow
	p.uiList.BorderLabel = fmt.Sprintf("%s: %s", p.ActiveQuery.Name, p.ActiveQuery.JQL)
	p.uiList.Height = ui.TermHeight() - 2
	p.uiList.Width = ui.TermWidth()
	p.uiList.Y = 0
	p.statusBar.Create()
	p.commandBar.Create()
	p.Update()
}
