'use strict';

// ── State ──────────────────────────────────────────────────────────────────
let state = {
  teams: [],
  dienste: [],
  persons: [],
  einsaetze: [],
  teamleiter: {},
  masterData: null,
  factDefs: [],
  _pickerCallback: null,
  _pickerCurrentId: '',
};

// ── Tab routing ────────────────────────────────────────────────────────────
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    const tab = btn.dataset.tab;
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-panel').forEach(p => {
      p.classList.remove('active');
      p.classList.add('hidden');
    });
    btn.classList.add('active');
    const panel = document.getElementById('tab-' + tab);
    panel.classList.remove('hidden');
    panel.classList.add('active');

    if (tab === 'musikteam') renderMusikteam();
    if (tab === 'einsaetze') renderEinsaetze();
    if (tab === 'teamleiter') renderTeamleiter();
  });
});

// ── Init ───────────────────────────────────────────────────────────────────
window.addEventListener('load', async () => {
  await loadSettings();
  await renderMusikteam();
  await loadEinsaetze();
  await loadTeamleiter();

  window.runtime.EventsOn('run:log', msg => {
    const log = document.getElementById('run-log');
    log.textContent += msg + '\n';
    log.scrollTop = log.scrollHeight;
  });
  window.runtime.EventsOn('run:done', result => {
    const log = document.getElementById('run-log');
    log.textContent += '\n' + '='.repeat(60) + '\n';
    log.textContent += 'Zusammenfassung:\n';
    log.textContent += `  Erfolgreich aktualisiert: ${result.updated} Dienstanfragen\n`;
    log.textContent += `  Fehler: ${result.errors}\n`;
    log.textContent += '='.repeat(60) + '\n';
    log.scrollTop = log.scrollHeight;
    document.getElementById('run-btn').disabled = false;
  });
});

// ── SETTINGS ──────────────────────────────────────────────────────────────
async function loadSettings() {
  const s = await window.go.main.App.GetSettings();
  document.getElementById('ct-url').value = s.ctURL || '';
  document.getElementById('api-key').value = s.apiKey || '';
  document.getElementById('store-key-cb').checked = s.storeAPIKey !== false;
  updateCtLink(s.ctURL);
  updateUrlPrompt(s.ctURL);
  if (s.apiKey) showPrivacyWarning();
  // First run: open settings automatically when no URL is configured
  if (!(s.ctURL || '').trim()) {
    document.getElementById('gear-btn').click();
  }
}

function updateUrlPrompt(url) {
  const isEmpty = !(url || '').trim();
  document.getElementById('url-welcome').classList.toggle('hidden', !isEmpty);
  document.getElementById('ct-url').classList.toggle('needs-input', isEmpty);
  document.getElementById('gear-btn').classList.toggle('needs-attention', isEmpty);
}

document.getElementById('ct-url').addEventListener('input', e => {
  updateCtLink(e.target.value);
  updateUrlPrompt(e.target.value);
});
document.getElementById('ct-url').addEventListener('blur', () => {
  const ctURL = document.getElementById('ct-url').value.trim();
  const apiKey = document.getElementById('api-key').value.trim();
  const storeKey = document.getElementById('store-key-cb').checked;
  window.go.main.App.SaveSettings(ctURL, apiKey, storeKey).catch(() => {});
});
document.getElementById('api-key').addEventListener('input', e => {
  if (e.target.value.trim()) showPrivacyWarning();
  else document.getElementById('privacy-warning').classList.add('hidden');
});

function updateCtLink(url) {
  url = (url || '').trim().replace(/\/$/, '');
  const link = document.getElementById('ct-link');
  if (url) {
    link.href = url;
    link.textContent = url;
  } else {
    link.href = '#';
    link.textContent = 'deine ChurchTools-URL';
  }
}

function showPrivacyWarning() {
  document.getElementById('privacy-warning').classList.remove('hidden');
}

function toggleKeyVisibility() {
  const input = document.getElementById('api-key');
  const btn = document.getElementById('toggle-key-btn');
  if (input.type === 'password') {
    input.type = 'text';
    btn.textContent = 'Verbergen';
  } else {
    input.type = 'password';
    btn.textContent = 'Anzeigen';
  }
}

async function saveSettings() {
  const ctURL = document.getElementById('ct-url').value.trim();
  const apiKey = document.getElementById('api-key').value.trim();
  const storeKey = document.getElementById('store-key-cb').checked;
  try {
    await window.go.main.App.SaveSettings(ctURL, apiKey, storeKey);
    await loadSettings(); // ctURL may have been normalized (e.g. https:// added) — reflect that in the field
    document.querySelector('.tab-btn[data-tab="musikteam"]').click();
  } catch (e) {
    showStatus('settings-status', 'Fehler: ' + e, true);
  }
}

// ── MUSIKTEAM ─────────────────────────────────────────────────────────────
async function loadMusikteam() {
  const data = await window.go.main.App.GetMusikteam();
  state.teams = data.teams || [];
  state.dienste = data.dienste || [];
  state.persons = await window.go.main.App.GetPersons();
  updateSetupGuide();
}

// Render without re-fetching from backend (use current state)
function renderMusikteamTable() {
  const header = document.getElementById('musikteam-header');
  const body = document.getElementById('musikteam-body');

  header.innerHTML =
    '<th>Bezeichnung</th><th>Dienst-ID</th>' +
    state.teams.map(t => `<th>${escHtml(t)}</th>`).join('') +
    '<th></th>';

  body.innerHTML = '';
  state.dienste.forEach((dienst, row) => {
    const tr = document.createElement('tr');

    // Label
    tr.appendChild(makeEditCell(dienst.label, val => {
      state.dienste[row].label = val;
      saveMusikteam();
    }));

    // ID
    tr.appendChild(makeEditCell(dienst.id, val => {
      state.dienste[row].id = val;
      saveMusikteam();
    }));

    // Team cells
    state.teams.forEach(team => {
      const td = document.createElement('td');
      const personId = (dienst.persons || {})[team] || '';
      td.appendChild(makeTeamCell(personId, team, row));
      tr.appendChild(td);
    });

    // Delete button
    const tdDel = document.createElement('td');
    const delBtn = document.createElement('button');
    delBtn.className = 'row-delete-btn';
    delBtn.textContent = 'x';
    delBtn.title = 'Dienst entfernen';
    delBtn.addEventListener('click', () => removeDienstAt(row));
    tdDel.appendChild(delBtn);
    tr.appendChild(tdDel);

    body.appendChild(tr);
  });
}

// Full render: reload from backend then repaint
async function renderMusikteam() {
  await loadMusikteam();
  renderMusikteamTable();
}

function makeEditCell(value, onChange) {
  const td = document.createElement('td');
  const input = document.createElement('input');
  input.type = 'text';
  input.value = value || '';
  input.addEventListener('change', () => onChange(input.value.trim()));
  td.appendChild(input);
  return td;
}

function makeTeamCell(personId, team, row) {
  const div = document.createElement('div');
  div.className = 'team-cell' + (personId ? ' filled' : '');

  const name = document.createElement('span');
  name.className = 'person-name';
  name.textContent = personId ? resolvePersonName(personId) : 'Klicken...';

  const icon = document.createElement('span');
  icon.className = 'edit-icon';
  icon.textContent = 'bearbeiten';

  div.appendChild(name);
  div.appendChild(icon);

  div.addEventListener('click', () => {
    if (state.persons.length === 0) {
      alert('Bitte zuerst Personen aus ChurchTools laden (Button „Personen laden").');
      return;
    }
    openPersonModal(personId, async selectedId => {
      if (!state.dienste[row].persons) state.dienste[row].persons = {};
      if (selectedId) {
        state.dienste[row].persons[team] = selectedId;
        div.classList.add('filled');
        name.textContent = resolvePersonName(selectedId);
        personId = selectedId;
      } else {
        delete state.dienste[row].persons[team];
        div.classList.remove('filled');
        name.textContent = 'Klicken...';
        personId = '';
      }
      await saveMusikteam();
      showStatus('persons-status', 'Gespeichert.');
    });
  });

  return div;
}

function resolvePersonName(id) {
  const p = state.persons.find(p => String(p.id) === String(id));
  return p ? p.name : `ID: ${id}`;
}

function updateSetupGuide() {
  const diensteDone = state.dienste.length > 0;
  const teamsDone = state.teams.length > 0;
  const personsDone = state.persons.length > 0;
  const allDone = diensteDone && teamsDone && personsDone;

  document.getElementById('setup-step-dienste').classList.toggle('done', diensteDone);
  document.getElementById('setup-step-teams').classList.toggle('done', teamsDone);
  document.getElementById('setup-step-persons').classList.toggle('done', personsDone);

  document.getElementById('musikteam-setup').classList.toggle('hidden', allDone);
  document.getElementById('musikteam-toolbar').classList.toggle('hidden', !allDone);

  updatePersonsStatus();
}

// ── ADD TEAM MODAL ─────────────────────────────────────────────────────────
function addTeam() {
  document.getElementById('add-team-count').value = '1';
  updateAddTeamInputs();
  document.getElementById('add-team-modal').classList.remove('hidden');
}

function updateAddTeamInputs() {
  const count = Math.max(1, Math.min(20, parseInt(document.getElementById('add-team-count').value) || 1));
  const container = document.getElementById('add-team-names');
  container.innerHTML = '';
  for (let i = 0; i < count; i++) {
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'add-team-name-input';
    input.placeholder = `Team-Name ${i + 1}`;
    container.appendChild(input);
  }
}

async function confirmAddTeams() {
  const inputs = document.querySelectorAll('.add-team-name-input');
  const toAdd = [];
  inputs.forEach(inp => {
    const t = inp.value.trim();
    if (!t) return;
    if (state.teams.includes(t)) { inp.style.borderColor = '#ff3b30'; return; }
    toAdd.push(t);
  });
  if (toAdd.length === 0) return;
  toAdd.forEach(t => {
    state.teams.push(t);
    state.dienste.forEach(d => { if (!d.persons) d.persons = {}; });
  });
  closeAddTeamModal();
  await saveMusikteam();
  updateSetupGuide();
  renderMusikteamTable();
}

function closeAddTeamModal() {
  document.getElementById('add-team-modal').classList.add('hidden');
}

document.getElementById('add-team-modal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeAddTeamModal();
});

// ── REMOVE TEAM MODAL ──────────────────────────────────────────────────────
function removeTeam() {
  const list = document.getElementById('remove-team-list');
  list.innerHTML = '';
  if (state.teams.length === 0) {
    const li = document.createElement('li');
    li.textContent = 'Keine Teams vorhanden.';
    li.style.padding = '10px 14px';
    li.style.color = '#6e6e73';
    list.appendChild(li);
  } else {
    state.teams.forEach(t => {
      const li = document.createElement('li');
      li.className = 'remove-team-item';
      const name = document.createElement('span');
      name.textContent = t;
      const btn = document.createElement('button');
      btn.className = 'row-delete-btn';
      btn.textContent = 'Entfernen';
      btn.addEventListener('click', async () => {
        state.teams = state.teams.filter(x => x !== t);
        state.dienste.forEach(d => { if (d.persons) delete d.persons[t]; });
        await saveMusikteam();
        updateSetupGuide();
        renderMusikteamTable();
        removeTeam(); // refresh modal list
        if (state.teams.length === 0) closeRemoveTeamModal();
      });
      li.appendChild(name);
      li.appendChild(btn);
      list.appendChild(li);
    });
  }
  document.getElementById('remove-team-modal').classList.remove('hidden');
}

function closeRemoveTeamModal() {
  document.getElementById('remove-team-modal').classList.add('hidden');
}

document.getElementById('remove-team-modal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeRemoveTeamModal();
});

async function removeDienstAt(row) {
  state.dienste.splice(row, 1);
  await saveMusikteam();
  renderMusikteamTable();
}

async function saveMusikteam() {
  try {
    await window.go.main.App.SaveMusikteam(state.teams, state.dienste);
  } catch (e) {
    console.error('saveMusikteam:', e);
  }
}

async function fetchPersons() {
  const btn = document.getElementById('fetch-btn');
  const status = document.getElementById('persons-status');
  btn.disabled = true;
  status.textContent = 'Lade Personen...';
  try {
    state.persons = await window.go.main.App.FetchPersons();
    updateSetupGuide();
    renderMusikteamTable();
  } catch (e) {
    status.textContent = 'Fehler: ' + e;
  } finally {
    btn.disabled = false;
  }
}

function updatePersonsStatus() {
  const el = document.getElementById('persons-status');
  if (!el) return;
  el.textContent = state.persons.length
    ? `${state.persons.length} Personen im Cache`
    : '';
}

// ── SERVICE GROUP MODAL ────────────────────────────────────────────────────
async function loadDiensteFromCT() {
  const btn = document.getElementById('load-dienste-btn');
  btn.disabled = true;
  btn.textContent = 'Lade...';
  try {
    state.masterData = await window.go.main.App.FetchMasterData();
    openSGModal();
  } catch (e) {
    alert('Fehler beim Laden der Masterdata: ' + e);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Dienste aus CT laden';
  }
}

function openSGModal() {
  const md = state.masterData;
  const list = document.getElementById('sg-list');
  list.innerHTML = '';

  md.serviceGroups.forEach(sg => {
    const services = md.services.filter(s => s.serviceGroupId === sg.id);
    if (services.length === 0) return;

    // Group header with "select all" checkbox
    const groupHeader = document.createElement('div');
    groupHeader.className = 'sg-group-header';

    const groupLabel = document.createElement('label');
    groupLabel.className = 'sg-group-label';

    const groupCb = document.createElement('input');
    groupCb.type = 'checkbox';
    groupCb.className = 'sg-group-cb';
    groupCb.dataset.groupId = String(sg.id);

    const groupInfo = document.createElement('span');
    groupInfo.innerHTML = `<strong>${escHtml(sg.name)}</strong> <span class="sg-count">${services.length} Dienste</span>`;

    groupLabel.appendChild(groupCb);
    groupLabel.appendChild(groupInfo);
    groupHeader.appendChild(groupLabel);
    list.appendChild(groupHeader);

    // Individual services
    const servicesDiv = document.createElement('div');
    servicesDiv.className = 'sg-services';

    services.forEach(s => {
      const serviceRow = document.createElement('div');
      serviceRow.className = 'sg-row';

      const cb = document.createElement('input');
      cb.type = 'checkbox';
      cb.value = String(s.id);
      cb.className = 'sg-service-cb';
      cb.dataset.groupId = String(sg.id);

      const nameLabel = document.createElement('label');
      nameLabel.className = 'sg-service-name';
      nameLabel.textContent = s.name;
      nameLabel.addEventListener('click', () => { cb.checked = !cb.checked; cb.dispatchEvent(new Event('change')); });

      const countInput = document.createElement('input');
      countInput.type = 'number';
      countInput.min = '1';
      countInput.max = '10';
      countInput.value = '1';
      countInput.className = 'sg-count-input hidden';
      countInput.addEventListener('click', e => e.stopPropagation());

      serviceRow.appendChild(cb);
      serviceRow.appendChild(nameLabel);
      serviceRow.appendChild(countInput);
      servicesDiv.appendChild(serviceRow);

      cb.addEventListener('change', () => {
        countInput.classList.toggle('hidden', !cb.checked);
        updateGroupCbState(sg.id, servicesDiv, groupCb);
      });
    });

    list.appendChild(servicesDiv);

    groupCb.addEventListener('change', () => {
      servicesDiv.querySelectorAll('.sg-service-cb').forEach(cb => { cb.checked = groupCb.checked; });
    });
  });

  document.getElementById('sg-modal').classList.remove('hidden');
}

function updateGroupCbState(groupId, servicesDiv, groupCb) {
  const cbs = Array.from(servicesDiv.querySelectorAll('.sg-service-cb'));
  const allChecked = cbs.every(cb => cb.checked);
  const someChecked = cbs.some(cb => cb.checked);
  groupCb.checked = allChecked;
  groupCb.indeterminate = someChecked && !allChecked;
}

async function confirmServiceGroups() {
  const checkedCbs = document.querySelectorAll('#sg-list .sg-service-cb:checked');
  let added = 0;

  checkedCbs.forEach(cb => {
    const sid = cb.value;
    const countInput = cb.closest('.sg-row').querySelector('.sg-count-input');
    const count = Math.max(1, parseInt(countInput ? countInput.value : '1') || 1);
    const s = state.masterData.services.find(sv => String(sv.id) === sid);
    if (!s) return;

    for (let i = 0; i < count; i++) {
      const label = count > 1 ? `${s.name} (${i + 1})` : s.name;
      const persons = {};
      state.teams.forEach(t => { persons[t] = ''; });
      state.dienste.push({ id: sid, label, persons });
      added++;
    }
  });

  closeSGModal();
  if (added === 0) { alert('Keine Dienste ausgewählt.'); return; }
  await saveMusikteam();
  updateSetupGuide();
  renderMusikteamTable();
}

function closeSGModal() {
  document.getElementById('sg-modal').classList.add('hidden');
}

document.getElementById('sg-modal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeSGModal();
});

// ── EINSAETZE ─────────────────────────────────────────────────────────────
async function loadEinsaetze() {
  state.einsaetze = await window.go.main.App.GetEinsaetze() || [];
}

async function renderEinsaetze() {
  await loadEinsaetze();
  const data = await window.go.main.App.GetMusikteam();
  state.teams = data.teams || [];

  // Pre-fill date range inputs on first load
  const fromInput = document.getElementById('event-from');
  const toInput = document.getElementById('event-to');
  if (!fromInput.value) {
    const now = new Date();
    fromInput.value = now.toISOString().substring(0, 10);
    const future = new Date(now);
    future.setMonth(future.getMonth() + 3);
    toInput.value = future.toISOString().substring(0, 10);
  }

  const body = document.getElementById('einsaetze-body');
  body.innerHTML = '';

  state.einsaetze.forEach((e, i) => {
    const tr = document.createElement('tr');

    // Date
    const tdDate = document.createElement('td');
    const inputDate = document.createElement('input');
    inputDate.type = 'text';
    inputDate.value = e.datum || '';
    inputDate.addEventListener('change', () => {
      const val = inputDate.value.trim();
      if (val && !isValidDate(val)) {
        inputDate.classList.add('invalid');
      } else {
        inputDate.classList.remove('invalid');
        state.einsaetze[i].datum = val;
        saveEinsaetze();
      }
    });
    tdDate.appendChild(inputDate);
    tr.appendChild(tdDate);

    // Zeit (read-only)
    const tdZeit = document.createElement('td');
    tdZeit.className = 'event-name-cell';
    tdZeit.textContent = e.zeit || '–';
    tr.appendChild(tdZeit);

    // Event name (read-only)
    const tdEvent = document.createElement('td');
    tdEvent.className = 'event-name-cell';
    tdEvent.textContent = e.event_name || '–';
    tr.appendChild(tdEvent);

    // Team select
    const tdTeam = document.createElement('td');
    const sel = document.createElement('select');
    state.teams.forEach(t => {
      const opt = document.createElement('option');
      opt.value = t;
      opt.textContent = t;
      if (t === e.musikteam) opt.selected = true;
      sel.appendChild(opt);
    });
    sel.addEventListener('change', () => {
      state.einsaetze[i].musikteam = sel.value;
      saveEinsaetze();
    });
    tdTeam.appendChild(sel);
    tr.appendChild(tdTeam);

    // Delete
    const tdDel = document.createElement('td');
    const delBtn = document.createElement('button');
    delBtn.className = 'row-delete-btn';
    delBtn.textContent = 'x';
    delBtn.addEventListener('click', async () => {
      state.einsaetze.splice(i, 1);
      await saveEinsaetze();
      renderEinsaetze();
    });
    tdDel.appendChild(delBtn);
    tr.appendChild(tdDel);

    body.appendChild(tr);
  });
}

// ── EVENT PICKER ───────────────────────────────────────────────────────────
async function loadEventsForPicker() {
  const from = document.getElementById('event-from').value;
  const to = document.getElementById('event-to').value;
  if (!from || !to) { alert('Bitte Von- und Bis-Datum auswählen.'); return; }

  const btn = document.getElementById('load-events-btn');
  btn.disabled = true;
  btn.textContent = 'Lade...';
  try {
    const events = await window.go.main.App.FetchEvents(from, to);
    renderEventList(events);
  } catch (e) {
    alert('Fehler beim Laden der Events: ' + e);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Laden';
  }
}

function renderEventList(events) {
  const list = document.getElementById('event-list');
  const addRow = document.getElementById('add-events-row');
  list.innerHTML = '';

  if (!events || events.length === 0) {
    list.textContent = 'Keine Events im gewählten Zeitraum gefunden.';
    list.classList.remove('hidden');
    addRow.classList.add('hidden');
    return;
  }

  events.forEach(ev => {
    let isoDate = '', displayDate = '', zeit = '';
    if (ev.startDate) {
      const dt = new Date(ev.startDate);
      const y = dt.getFullYear();
      const mo = String(dt.getMonth() + 1).padStart(2, '0');
      const d = String(dt.getDate()).padStart(2, '0');
      isoDate = `${y}-${mo}-${d}`;
      displayDate = `${d}.${mo}.${y}`;
      zeit = `${String(dt.getHours()).padStart(2, '0')}:${String(dt.getMinutes()).padStart(2, '0')}`;
    }

    const row = document.createElement('div');
    row.className = 'event-row';

    const cb = document.createElement('input');
    cb.type = 'checkbox';
    cb.className = 'event-cb';
    cb.dataset.eventId = String(ev.id);
    cb.dataset.eventName = ev.name || '';
    cb.dataset.datum = displayDate;
    cb.dataset.zeit = zeit;

    const dateSpan = document.createElement('span');
    dateSpan.className = 'event-row-date';
    dateSpan.textContent = displayDate + (zeit ? ' ' + zeit : '');

    const nameSpan = document.createElement('span');
    nameSpan.className = 'event-row-name';
    nameSpan.textContent = ev.name || '';

    const teamSel = document.createElement('select');
    teamSel.className = 'event-row-team';
    const emptyOpt = document.createElement('option');
    emptyOpt.value = '';
    emptyOpt.textContent = 'Team wählen...';
    teamSel.appendChild(emptyOpt);
    (state.teams || []).forEach(t => {
      const opt = document.createElement('option');
      opt.value = t;
      opt.textContent = t;
      teamSel.appendChild(opt);
    });

    row.appendChild(cb);
    row.appendChild(dateSpan);
    row.appendChild(nameSpan);
    row.appendChild(teamSel);
    list.appendChild(row);
  });

  list.classList.remove('hidden');
  addRow.classList.remove('hidden');
}

async function addSelectedEvents() {
  const checkedCbs = document.querySelectorAll('#event-list .event-cb:checked');
  const missing = [];
  const toAdd = [];

  checkedCbs.forEach(cb => {
    const teamSel = cb.closest('.event-row').querySelector('.event-row-team');
    const team = teamSel ? teamSel.value : '';
    if (!team) { missing.push(cb.dataset.eventName || cb.dataset.datum); return; }
    toAdd.push({
      datum: cb.dataset.datum,
      zeit: cb.dataset.zeit || '',
      musikteam: team,
      event_id: parseInt(cb.dataset.eventId) || 0,
      event_name: cb.dataset.eventName,
    });
  });

  if (missing.length > 0) {
    alert('Bitte Team auswählen für:\n' + missing.join('\n'));
    return;
  }
  if (toAdd.length === 0) { alert('Keine Events ausgewählt.'); return; }

  toAdd.forEach(e => state.einsaetze.push(e));
  await saveEinsaetze();
  renderEinsaetze();
  // Hide the picker list after adding
  document.getElementById('event-list').classList.add('hidden');
  document.getElementById('add-events-row').classList.add('hidden');
}

async function addManualEinsatz() {
  const today = formatDateDE(new Date());
  const team = state.teams.length > 0 ? state.teams[0] : '';
  state.einsaetze.push({ datum: today, musikteam: team, event_id: 0, event_name: '' });
  await saveEinsaetze();
  renderEinsaetze();
}

async function saveEinsaetze() {
  try {
    await window.go.main.App.SetEinsaetze(state.einsaetze);
  } catch (e) {
    console.error('saveEinsaetze:', e);
  }
}

// ── TEAMLEITER ─────────────────────────────────────────────────────────────
async function loadTeamleiter() {
  state.teamleiter = await window.go.main.App.GetTeamleiter() || {};
}

async function loadFactsForTeamleiter() {
  const btn = document.getElementById('load-facts-btn');
  const status = document.getElementById('facts-status');
  btn.disabled = true;
  status.textContent = 'Lade...';
  try {
    state.factDefs = (await window.go.main.App.FetchFacts()) || [];
    await loadTeamleiter();
    const data = await window.go.main.App.GetMusikteam();
    state.teams = data.teams || [];
    renderTeamleiterTable();
    status.textContent = state.factDefs.length + ' Fakten geladen.';
    setTimeout(() => { status.textContent = ''; }, 3000);
  } catch (e) {
    status.textContent = 'Fehler: ' + e;
  } finally {
    btn.disabled = false;
  }
}

async function renderTeamleiter() {
  await loadTeamleiter();
  const data = await window.go.main.App.GetMusikteam();
  state.teams = data.teams || [];
  renderTeamleiterTable();
}

function renderTeamleiterTable() {
  const teams = state.teams;
  const facts = state.factDefs || [];
  const body = document.getElementById('teamleiter-body');
  body.innerHTML = '';

  teams.forEach(team => {
    // Always read from state.teamleiter fresh to get an up-to-date entry object
    if (!state.teamleiter[team]) state.teamleiter[team] = { fact_id: 0, value: '' };
    const entry = state.teamleiter[team];
    const tr = document.createElement('tr');

    // Team name
    const tdTeam = document.createElement('td');
    tdTeam.textContent = team;
    tr.appendChild(tdTeam);

    // Fact dropdown
    const tdFact = document.createElement('td');
    const sel = document.createElement('select');
    const emptyOpt = document.createElement('option');
    emptyOpt.value = '0';
    emptyOpt.textContent = facts.length === 0 ? 'Zuerst Fakten laden...' : '– Kein Fakt –';
    sel.appendChild(emptyOpt);
    facts.forEach(f => {
      const opt = document.createElement('option');
      opt.value = String(f.id);
      opt.textContent = f.nameTranslated || f.name || String(f.id);
      sel.appendChild(opt);
    });
    // Set the selected value after all options are built (more reliable than opt.selected = true)
    sel.value = String(entry.fact_id || 0);

    // Value cell — rebuilt whenever the fact selection changes
    const tdValue = document.createElement('td');

    function buildValueInput() {
      tdValue.innerHTML = '';
      const selectedFact = facts.find(f => f.id === entry.fact_id);
      if (selectedFact && selectedFact.options && selectedFact.options.length > 0) {
        // Select-type fact: show a dropdown with the predefined allowed values
        const valSel = document.createElement('select');
        const placeholder = document.createElement('option');
        placeholder.value = '';
        placeholder.textContent = '– Option wählen –';
        valSel.appendChild(placeholder);
        selectedFact.options.forEach(o => {
          const opt = document.createElement('option');
          opt.value = o;
          opt.textContent = o;
          valSel.appendChild(opt);
        });
        valSel.value = entry.value || '';
        valSel.addEventListener('change', () => {
          entry.value = valSel.value;
          saveTeamleiter();
        });
        tdValue.appendChild(valSel);
      } else {
        // Free-text fact
        const input = document.createElement('input');
        input.type = 'text';
        input.value = entry.value || '';
        input.placeholder = team;
        input.addEventListener('change', () => {
          entry.value = input.value.trim();
          saveTeamleiter();
        });
        tdValue.appendChild(input);
      }
    }

    sel.addEventListener('change', () => {
      entry.fact_id = parseInt(sel.value) || 0;
      buildValueInput();
      saveTeamleiter();
    });
    tdFact.appendChild(sel);
    tr.appendChild(tdFact);

    buildValueInput();
    tr.appendChild(tdValue);

    body.appendChild(tr);
  });
}

async function saveTeamleiter() {
  try {
    await window.go.main.App.SetTeamleiter(state.teamleiter);
  } catch (e) {
    console.error('saveTeamleiter:', e);
  }
}

// ── RUN ───────────────────────────────────────────────────────────────────
async function runUpdate() {
  const btn = document.getElementById('run-btn');
  const log = document.getElementById('run-log');
  btn.disabled = true;
  log.textContent = 'Starte Ausführung...\n';
  try {
    await window.go.main.App.RunUpdate();
  } catch (e) {
    log.textContent += 'FEHLER: ' + e + '\n';
    btn.disabled = false;
  }
}

// ── PERSON MODAL ───────────────────────────────────────────────────────────
function openPersonModal(currentId, callback) {
  state._pickerCallback = callback;
  state._pickerCurrentId = String(currentId || '');

  document.getElementById('person-search').value = '';
  filterPersons('');

  document.getElementById('person-modal').classList.remove('hidden');
  document.getElementById('person-search').focus();
}

function filterPersons(query) {
  query = query.toLowerCase();
  const list = document.getElementById('person-list');
  list.innerHTML = '';

  state.persons
    .filter(p => p.name.toLowerCase().includes(query))
    .forEach(p => {
      const li = document.createElement('li');
      li.dataset.id = String(p.id);
      li.textContent = p.name + '  (ID: ' + p.id + ')';
      if (String(p.id) === state._pickerCurrentId) li.classList.add('selected');
      li.addEventListener('click', () => {
        document.querySelectorAll('#person-list li').forEach(x => x.classList.remove('selected'));
        li.classList.add('selected');
      });
      li.addEventListener('dblclick', confirmPersonSelection);
      list.appendChild(li);
    });
}

function confirmPersonSelection() {
  const selected = document.querySelector('#person-list li.selected');
  const id = selected ? selected.dataset.id : '';
  const cb = state._pickerCallback;
  closePersonModal();
  if (cb) cb(id);
}

function clearPersonSelection() {
  const cb = state._pickerCallback;
  closePersonModal();
  if (cb) cb('');
}

function closePersonModal() {
  document.getElementById('person-modal').classList.add('hidden');
  state._pickerCallback = null;
}

document.getElementById('person-modal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closePersonModal();
});

// ── Helpers ────────────────────────────────────────────────────────────────
function isValidDate(s) {
  return /^\d{2}\.\d{2}\.\d{4}$/.test(s) && !isNaN(Date.parse(s.split('.').reverse().join('-')));
}

function formatDateDE(d) {
  return `${String(d.getDate()).padStart(2,'0')}.${String(d.getMonth()+1).padStart(2,'0')}.${d.getFullYear()}`;
}

function escHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function showStatus(id, msg, isError = false) {
  const el = document.getElementById(id);
  if (!el) return;
  el.textContent = msg;
  el.style.color = isError ? '#ff3b30' : '#34c759';
  setTimeout(() => { el.textContent = ''; el.style.color = ''; }, 3000);
}
