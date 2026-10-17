import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const EXTRACT_ROUTE = '**/api/campaigns/*/copilot/transcript/*/extract';

const EXTRACT_200 = {
  recap: { title: 'Session 5 Recap', content: 'The party delved deep.' },
  drafts: [{ id: 'd1', entity_type: 'npc', name: 'Grimble' }],
  source: { entity_type: 'wiki', entity_id: 1, title: 'Transcript' },
};

// Seeds a campaign, opens the copilot panel and provides a transcript id so the
// Recap + extract action becomes enabled.
async function openCopilotWithTranscript(page: any): Promise<void> {
  await login(page);
  const campId: number = await page.evaluate(async () => {
    const camp = await (window as any).api('POST', '/api/campaigns', {
      name: 'Extract Camp ' + Date.now(),
      description: 'extract e2e',
      dm_notes: '',
    });
    return camp.id;
  });
  await page.evaluate((cid: number) => (window as any).showCopilot(cid), campId);
  await expect(page.getByTestId('copilot-view')).toBeVisible({ timeout: NAV_TIMEOUT });
  await page.evaluate(() => {
    (window as any).__copilotTranscriptId = 1;
    (window as any).refreshCopilotPanel?.();
  });
  await expect(page.getByTestId('copilot-generate-recap-extract')).toBeVisible({ timeout: NAV_TIMEOUT });
}

test.describe('copilot Recap + extract', () => {
  test('action is present and enabled once a transcript exists', async ({ page }) => {
    await openCopilotWithTranscript(page);
    await expect(page.getByTestId('copilot-generate-recap-extract')).toBeEnabled({ timeout: NAV_TIMEOUT });
  });

  test('200 renders the recap preview and one extracted draft', async ({ page }) => {
    await openCopilotWithTranscript(page);
    let method = '';
    let url = '';
    await page.route(EXTRACT_ROUTE, async (route: any) => {
      method = route.request().method();
      url = route.request().url();
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(EXTRACT_200) });
    });

    await page.getByTestId('copilot-generate-recap-extract').click();

    await expect(page.getByTestId('copilot-recap-preview')).toContainText('Session 5 Recap', { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-extracted-drafts')).toContainText('Grimble');
    await expect(page.getByTestId('copilot-extracted-draft')).toHaveCount(1);
    await expect(page.getByTestId('copilot-draft-commit')).toBeVisible();
    await expect(page.getByTestId('copilot-draft-discard')).toBeVisible();
    expect(method).toBe('POST');
    expect(url).toContain('/extract');
  });

  test('commit issues a POST to the draft commit endpoint', async ({ page }) => {
    await openCopilotWithTranscript(page);
    await page.route(EXTRACT_ROUTE, (route: any) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(EXTRACT_200) })
    );
    let commitUrl = '';
    await page.route('**/api/ai/draft/*/commit', (route: any) => {
      commitUrl = route.request().url();
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: 'd1', entity_type: 'npc', entity_id: 9, url: '#/npcs/9' }),
      });
    });

    await page.getByTestId('copilot-generate-recap-extract').click();
    await page.getByTestId('copilot-draft-commit').click();

    await expect.poll(() => commitUrl, { timeout: NAV_TIMEOUT }).toContain('/api/ai/draft/d1/commit');
  });

  test('503 shows the returned error', async ({ page }) => {
    await openCopilotWithTranscript(page);
    await page.route(EXTRACT_ROUTE, (route: any) =>
      route.fulfill({ status: 503, contentType: 'application/json', body: JSON.stringify({ error: 'AI is not enabled' }) })
    );

    await page.getByTestId('copilot-generate-recap-extract').click();

    await expect(page.locator('#copilotExtractError')).toContainText('AI is not enabled', { timeout: NAV_TIMEOUT });
  });

  test('422 shows the error and message', async ({ page }) => {
    await openCopilotWithTranscript(page);
    await page.route(EXTRACT_ROUTE, (route: any) =>
      route.fulfill({
        status: 422,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'could not extract a recap or entities', message: 'no JSON' }),
      })
    );

    await page.getByTestId('copilot-generate-recap-extract').click();

    const err = page.locator('#copilotExtractError');
    await expect(err).toContainText('could not extract', { timeout: NAV_TIMEOUT });
    await expect(err).toContainText('no JSON');
  });
});
