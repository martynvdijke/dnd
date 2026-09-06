// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar } from '../lib/state';

// ─── D3 Force Graph ───

async function createForceGraph(
  container: HTMLElement,
  data: { nodes: any[], edges: any[] },
  groups: Record<string, { shape: string, color: string }>,
  options?: { linkDistance?: number, chargeStrength?: number }
) {
  let d3: any;
  try { d3 = await import('d3'); } catch (e) { console.warn('d3 load failed', e); return null; }
  const width = container.clientWidth || 800;
  const height = container.clientHeight || 600;

  container.innerHTML = '';

  const svg = d3.select(container)
    .append('svg')
    .attr('width', width)
    .attr('height', height)
    .style('background', 'var(--parchment-light)')
    .style('cursor', 'grab')
    .style('border-radius', '4px')
    .style('display', 'block');

  const strokeColor = '#2c1810';
  const edgeColor = '#8b7355';

  svg.append('defs').append('marker')
    .attr('id', 'arrowhead')
    .attr('viewBox', '0 -5 10 10')
    .attr('refX', 20)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-5L10,0L0,5')
    .attr('fill', edgeColor);

  const g = svg.append('g');

  const zoom = d3.zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.1, 4])
    .on('zoom', (event) => g.attr('transform', event.transform));
  svg.call(zoom);

  const link = g.append('g')
    .selectAll<SVGLineElement, any>('line')
    .data(data.edges)
    .join('line')
    .attr('stroke', edgeColor)
    .attr('stroke-width', (d: any) => d.width || 1)
    .attr('stroke-dasharray', (d: any) => d.dashes ? '6,3' : null)
    .attr('marker-end', 'url(#arrowhead)');

  const linkLabel = g.append('g')
    .selectAll<SVGTextElement, any>('text')
    .data(data.edges.filter((d: any) => d.label))
    .join('text')
    .text((d: any) => d.label)
    .attr('font-size', 10)
    .attr('font-family', 'Vollkorn')
    .attr('fill', '#5c3a2a')
    .attr('text-anchor', 'middle')
    .attr('dy', '-4');

  const node = g.append('g')
    .selectAll<SVGGElement, any>('g')
    .data(data.nodes)
    .join('g')
    .style('cursor', 'pointer');

  node.each(function (d: any) {
    const el = d3.select(this);
    const size = d.size || 15;
    const grp = groups[d.group] || { shape: 'dot', color: '#8b0000' };
    const color = d.color || grp.color;

    const shapeEl = (() => {
      switch (grp.shape) {
        case 'ellipse':
          return el.append('ellipse').attr('rx', size).attr('ry', size * 0.7);
        case 'square':
          return el.append('rect').attr('x', -size).attr('y', -size)
            .attr('width', size * 2).attr('height', size * 2).attr('rx', 3);
        case 'diamond': {
          const pts = `0,-${size} ${size},0 0,${size} -${size},0`;
          return el.append('polygon').attr('points', pts);
        }
        case 'star': {
          const pts: string[] = [];
          for (let i = 0; i < 10; i++) {
            const r = i % 2 === 0 ? size : size * 0.4;
            const a = (i * Math.PI) / 5 - Math.PI / 2;
            pts.push(`${(r * Math.cos(a)).toFixed(1)},${(r * Math.sin(a)).toFixed(1)}`);
          }
          return el.append('polygon').attr('points', pts.join(' '));
        }
        case 'hexagon': {
          const pts: string[] = [];
          for (let i = 0; i < 6; i++) {
            const a = (i * Math.PI * 2) / 6 - Math.PI / 2;
            pts.push(`${(size * Math.cos(a)).toFixed(1)},${(size * Math.sin(a)).toFixed(1)}`);
          }
          return el.append('polygon').attr('points', pts.join(' '));
        }
        case 'triangle':
          return el.append('polygon')
            .attr('points', `0,-${size} ${(size * 0.866).toFixed(1)},${(size * 0.5).toFixed(1)} -${(size * 0.866).toFixed(1)},${(size * 0.5).toFixed(1)}`);
        default:
          return el.append('circle').attr('r', size * 0.5);
      }
    })();

    shapeEl
      .attr('fill', color)
      .attr('stroke', strokeColor)
      .attr('stroke-width', 2);

    const labelSize = d.size > 20 ? 14 : 11;
    const dy = grp.shape === 'dot' ? size * 0.5 + 14 : size + 10;

    el.append('text')
      .text(d.label)
      .attr('dy', dy)
      .attr('text-anchor', 'middle')
      .attr('fill', strokeColor)
      .attr('font-family', 'Playfair Display')
      .attr('font-size', labelSize);

    el.on('mouseenter', () => shapeEl.attr('stroke', '#b8963e').attr('stroke-width', 3))
      .on('mouseleave', () => shapeEl.attr('stroke', strokeColor).attr('stroke-width', 2));
  });

  const drag = d3.drag<SVGGElement, any>()
    .on('start', (event, d) => {
      if (!event.active) sim.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    })
    .on('drag', (event, d) => { d.fx = event.x; d.fy = event.y; })
    .on('end', (event, d) => {
      if (!event.active) sim.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    });

  node.call(drag as any);

  const sim = d3.forceSimulation(data.nodes)
    .force('link', d3.forceLink(data.edges.map((e: any) => ({ ...e, source: e.from, target: e.to })))
      .id((d: any) => d.id)
      .distance(options?.linkDistance || 200))
    .force('charge', d3.forceManyBody().strength(options?.chargeStrength || -300))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius((d: any) => d.size + 20))
    .on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);
      linkLabel
        .attr('x', (d: any) => (d.source.x + d.target.x) / 2)
        .attr('y', (d: any) => (d.source.y + d.target.y) / 2);
      node.attr('transform', (d: any) => `translate(${d.x},${d.y})`);
    });

  const ro = new ResizeObserver(() => {
    const w = container.clientWidth;
    const h = container.clientHeight;
    svg.attr('width', w).attr('height', h);
    sim.force('center', d3.forceCenter(w / 2, h / 2)).alpha(0.3).restart();
  });
  ro.observe(container);

  return sim;
}
expose('createForceGraph', createForceGraph);

// ─── Graph ───

async function renderGraph() {
  const el = document.getElementById('graphSection')!;
  el.innerHTML = `<div class="ornament mb-3">✧ Drawing your web of fate ✧</div>
    <div id="graphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>`;
  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/graph`);
    const container = document.getElementById('graphContainer')!;
    await createForceGraph(container, data, {
      character: { shape: 'ellipse', color: '#8b0000' },
      location: { shape: 'square', color: '#b8963e' },
      npc: { shape: 'diamond', color: '#2d6a2d' },
      quest: { shape: 'star', color: '#8b4513' },
      session: { shape: 'dot', color: '#5c3a2a' },
    }, { linkDistance: 200, chargeStrength: -300 });
  } catch (e:any) {
    el.innerHTML += `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load graph: ${esc(e.message)}</p></div>`;
  }
}
expose('renderGraph', renderGraph);

// ─── Analytics ───

async function renderAnalytics() {
  const el = document.getElementById('analyticsSection')!;
  el.innerHTML = '<div class="ornament mb-3">✧ Loading analytics... ✧</div>';
  try {
    const stats = await api('GET', `/api/characters/${currentChar.id}/stats`);
    el.innerHTML = `
      <h5>Campaign Overview</h5>
      <div class="row g-3 mb-3">
        <div class="col-6 col-md-3"><div class="combat-stat"><div class="stat-label">Sessions</div><div class="stat-value">${stats.session_count}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat"><div class="stat-label">Level</div><div class="stat-value">${stats.level}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat text-success"><div class="stat-label">Total XP</div><div class="stat-value">${stats.total_xp_earned}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat" style="color:var(--gold)"><div class="stat-label">Gold Earned</div><div class="stat-value">${stats.total_gold_earned}</div></div></div>
      </div>
      <div class="row g-3 mb-3">
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Quests (${stats.quests.total})</h6>
              <div class="d-flex gap-1 flex-wrap">
                ${stats.quests.active > 0 ? `<span class="badge badge-blood">${stats.quests.active} Active</span>` : ''}
                ${stats.quests.complete > 0 ? `<span class="badge bg-success">${stats.quests.complete} Complete</span>` : ''}
                ${stats.quests.failed > 0 ? `<span class="badge bg-secondary">${stats.quests.failed} Failed</span>` : ''}
                ${stats.quests.available > 0 ? `<span class="badge badge-gold">${stats.quests.available} Available</span>` : ''}
              </div>
            </div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Rests</h6>
              <div class="d-flex gap-1 flex-wrap">
                <span class="badge badge-gold">${stats.rests.short} Short</span>
                <span class="badge badge-blood">${stats.rests.long} Long</span>
                ${stats.rests.total_healed > 0 ? `<span class="badge bg-success">${stats.rests.total_healed} HP Healed</span>` : ''}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="row g-3 mb-3">
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>World</h6>
              <p class="mb-1 small text-muted">${stats.locations_count} Locations explored</p>
              <p class="mb-1 small text-muted">${stats.npc_interactions} NPC interactions</p>
              <p class="mb-1 small text-muted">${stats.journal_count} Journal entries</p>
              <p class="mb-0 small text-muted">${stats.dice_rolls.total_rolls} Dice rolls (avg ${stats.dice_rolls.average.toFixed(1)})</p>
            </div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Notable NPCs</h6>
              ${stats.top_npcs && stats.top_npcs.length > 0
                ? stats.top_npcs.map((n:any) => `<p class="mb-1 small text-muted">&loz; ${esc(n)}</p>`).join('')
                : '<p class="mb-0 small text-muted fst-italic">No NPC interactions yet</p>'}
            </div>
          </div>
        </div>
      </div>
      <div id="questChartContainer" style="height:200px;max-width:400px;margin:0 auto"></div>`;
    if (stats.quests.total > 0) {
      try {
        const { default: Chart } = await import('chart.js/auto');
        const ctx = document.createElement('canvas');
        document.getElementById('questChartContainer')!.appendChild(ctx);
        new Chart(ctx, {
          type: 'doughnut',
          data: {
            labels: ['Active', 'Complete', 'Failed', 'Available', 'Abandoned'],
            datasets: [{
              data: [stats.quests.active, stats.quests.complete, stats.quests.failed, stats.quests.available, stats.quests.abandoned],
              backgroundColor: ['#8b0000', '#2d6a2d', '#666', '#b8963e', '#ccc'],
              borderWidth: 0,
            }]
          },
          options: {
            responsive: true, maintainAspectRatio: false,
            plugins: { legend: { position: 'bottom', labels: { font: { family: 'Vollkorn' } } } }
          }
        });
      } catch (e) { console.warn('chart.js load failed', e); }
    }
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load analytics: ${esc(e.message)}</p></div>`;
  }
}
expose('renderAnalytics', renderAnalytics);
