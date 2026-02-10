(function () {
    var searchInput = document.getElementById('search-input');
    var searchResults = document.getElementById('search-results');
    var selectedCv = document.getElementById('selected-cv');
    var selectedName = document.getElementById('selected-name');
    var selectedLattesId = document.getElementById('selected-lattes-id');
    var spinner = document.getElementById('spinner');
    var errorMsg = document.getElementById('error-message');
    var infoMsg = document.getElementById('info-message');
    var analysisSection = document.getElementById('analysis-section');
    var metadata = document.getElementById('metadata');
    var analysisContent = document.getElementById('analysis-content');
    var downloadMd = document.getElementById('download-md');

    var currentLattesId = '';
    var currentAnalysis = '';
    var searchTimeout = null;

    searchInput.addEventListener('input', function () {
        var query = searchInput.value.trim();
        if (searchTimeout) clearTimeout(searchTimeout);

        if (query.length < 3) {
            searchResults.innerHTML = '';
            return;
        }

        searchTimeout = setTimeout(function () {
            fetch('/api/search?q=' + encodeURIComponent(query))
                .then(function (r) { return r.json(); })
                .then(function (data) {
                    if (!data.success || !data.results || data.results.length === 0) {
                        searchResults.innerHTML = '<p class="search-empty">Nenhum resultado encontrado</p>';
                        return;
                    }
                    var html = '';
                    for (var i = 0; i < data.results.length; i++) {
                        var cv = data.results[i];
                        html += '<div class="search-result-card" data-lattes-id="' + cv.lattesId + '" data-name="' + escapeHtml(cv.name) + '">';
                        html += '<strong>' + escapeHtml(cv.name) + '</strong>';
                        html += '<span class="search-result-id">' + cv.lattesId + '</span>';
                        html += '</div>';
                    }
                    searchResults.innerHTML = html;

                    var cards = searchResults.querySelectorAll('.search-result-card');
                    for (var j = 0; j < cards.length; j++) {
                        cards[j].addEventListener('click', function () {
                            selectCV(this.getAttribute('data-lattes-id'), this.getAttribute('data-name'));
                        });
                    }
                })
                .catch(function () {
                    searchResults.innerHTML = '<p class="search-empty">Erro ao buscar</p>';
                });
        }, 300);
    });

    function selectCV(lattesId, name) {
        currentLattesId = lattesId;
        selectedName.textContent = name;
        selectedLattesId.textContent = lattesId;
        selectedCv.style.display = 'block';
        searchResults.innerHTML = '';
        hideMessages();
        analysisSection.style.display = 'none';

        loadAnalysis(lattesId);
    }

    function loadAnalysis(lattesId) {
        spinner.classList.add('visible');
        hideMessages();

        fetch('/api/analysis/view/' + lattesId)
            .then(function (r) {
                return r.json().then(function (data) {
                    return { status: r.status, body: data };
                });
            })
            .then(function (result) {
                spinner.classList.remove('visible');

                if (result.status === 404) {
                    showInfo(result.body.error || 'Nenhuma análise salva para este pesquisador');
                    return;
                }

                if (!result.body.success) {
                    showError(result.body.error || 'Erro ao carregar análise');
                    return;
                }

                currentAnalysis = result.body.analysis;

                var metaHtml = '<p class="metadata-text">Gerado por <strong>' +
                    escapeHtml(result.body.provider) + '</strong> / <strong>' +
                    escapeHtml(result.body.model) + '</strong>';
                if (result.body.generatedAt) {
                    var date = new Date(result.body.generatedAt);
                    metaHtml += ' em ' + date.toLocaleDateString('pt-BR') + ' ' + date.toLocaleTimeString('pt-BR');
                }
                if (result.body.researchersAnalyzed) {
                    metaHtml += ' &mdash; ' + result.body.researchersAnalyzed + ' pesquisadores analisados';
                }
                metaHtml += '</p>';
                metadata.innerHTML = metaHtml;

                analysisContent.innerHTML = renderMarkdown(result.body.analysis);
                analysisSection.style.display = 'block';
            })
            .catch(function () {
                spinner.classList.remove('visible');
                showError('Erro de conexão ao carregar análise');
            });
    }

    downloadMd.addEventListener('click', function () {
        downloadBlob(currentAnalysis, 'analise-' + currentLattesId + '.md', 'text/markdown');
    });

    function showError(message) {
        errorMsg.textContent = message;
        errorMsg.classList.add('visible');
    }

    function showInfo(message) {
        infoMsg.textContent = message;
        infoMsg.classList.add('visible');
    }

    function hideMessages() {
        errorMsg.classList.remove('visible');
        infoMsg.classList.remove('visible');
    }

    function escapeHtml(text) {
        var div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function downloadBlob(content, filename, mimeType) {
        var blob = new Blob([content], { type: mimeType + '; charset=utf-8' });
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function renderMarkdown(md) {
        var html = md
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');

        html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>');
        html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>');

        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        html = html.replace(/^\|(.+)\|$/gm, function (match, content) {
            var cells = content.split('|').map(function (c) { return c.trim(); });
            return '<tr>' + cells.map(function (c) {
                if (/^[-:]+$/.test(c)) return '';
                return '<td>' + c + '</td>';
            }).join('') + '</tr>';
        });
        html = html.replace(/((?:<tr>.*?<\/tr>\n?)+)/g, '<table class="summary-table">$1</table>');
        html = html.replace(/<tr><\/tr>/g, '');

        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');
        html = html.replace(/((?:<li>.*?<\/li>\n?)+)/g, '<ul>$1</ul>');

        html = html.replace(/^(?!<[hultd])(.+)$/gm, '<p>$1</p>');
        html = html.replace(/<p>\s*<\/p>/g, '');

        return html;
    }
})();
