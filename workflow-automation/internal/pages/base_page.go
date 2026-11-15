package pages

import (
    "workflow-automation/internal/logger"
    "github.com/mxschmitt/playwright-go"
)

type BasePage struct {
    Page playwright.Page
}

func NewBasePage(page playwright.Page) BasePage {
    return BasePage{Page: page}
}

func (b *BasePage) Click(selector string) error {
    logger.Debug("BasePage:Click", "Clicking "+selector)
    return b.Page.Click(selector)
}

func (b *BasePage) Fill(selector, value string) error {
    logger.Debug("BasePage:Fill", "Filling "+selector)
    return b.Page.Fill(selector, value)
}

func (b *BasePage) WaitForNavigation() error {
    return b.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle})
}
