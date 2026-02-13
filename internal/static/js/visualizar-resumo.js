(function () {
    var searchInput = document.getElementById('search-input');
    var searchResults = document.getElementById('search-results');
    var selectedCv = document.getElementById('selected-cv');
    var selectedName = document.getElementById('selected-name');
    var selectedLattesId = document.getElementById('selected-lattes-id');
    var spinner = document.getElementById('spinner');
    var errorMsg = document.getElementById('error-message');
    var infoMsg = document.getElementById('info-message');
    var summarySection = document.getElementById('summary-section');
    var metadata = document.getElementById('metadata');
    var summaryContent = document.getElementById('summary-content');
    var downloadMd = document.getElementById('download-md');
    var downloadPdf = document.getElementById('download-pdf');

    var shareBtn = document.getElementById('share-btn');

    var currentLattesId = '';
    var currentSummary = '';
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
        summarySection.style.display = 'none';

        loadSummary(lattesId);
    }

    function loadSummary(lattesId) {
        spinner.classList.add('visible');
        hideMessages();

        fetch('/api/summary/view/' + lattesId)
            .then(function (r) {
                return r.json().then(function (data) {
                    return { status: r.status, body: data };
                });
            })
            .then(function (result) {
                spinner.classList.remove('visible');

                if (result.status === 404) {
                    showInfo(result.body.error || 'Nenhum resumo salvo para este pesquisador');
                    return;
                }

                if (!result.body.success) {
                    showError(result.body.error || 'Erro ao carregar resumo');
                    return;
                }

                currentSummary = result.body.summary;

                var metaHtml = '<p class="metadata-text">Gerado por <strong>' +
                    escapeHtml(result.body.provider) + '</strong> / <strong>' +
                    escapeHtml(result.body.model) + '</strong>';
                if (result.body.generatedAt) {
                    var date = new Date(result.body.generatedAt);
                    metaHtml += ' em ' + date.toLocaleDateString('pt-BR') + ' ' + date.toLocaleTimeString('pt-BR');
                }
                metaHtml += '</p>';
                metadata.innerHTML = metaHtml;

                summaryContent.innerHTML = renderMarkdown(result.body.summary);
                summarySection.style.display = 'block';

                if (shareBtn) shareBtn.style.display = '';
            })
            .catch(function () {
                spinner.classList.remove('visible');
                showError('Erro de conexão ao carregar resumo');
            });
    }

    downloadMd.addEventListener('click', function () {
        downloadBlob(currentSummary, 'resumo-' + currentLattesId + '.md', 'text/markdown');
    });
    downloadPdf.addEventListener('click', function () {
        downloadAsPdf(currentSummary);
    });

    if (shareBtn) {
        shareBtn.addEventListener('click', function () {
            fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
                var url = cfg.shareBaseUrl + '?resumo=' + currentLattesId;
                copyToClipboard(url, shareBtn);
            });
        });
    }

    function copyToClipboard(text, btn) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function () {
                showShareFeedback(btn);
            }).catch(function () {
                fallbackCopy(text, btn);
            });
        } else {
            fallbackCopy(text, btn);
        }
    }

    function fallbackCopy(text, btn) {
        var ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.left = '-9999px';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        showShareFeedback(btn);
    }

    function showShareFeedback(btn) {
        var original = btn.textContent;
        btn.textContent = 'Link copiado!';
        btn.disabled = true;
        setTimeout(function () { btn.textContent = original; btn.disabled = false; }, 2000);
    }

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

    function markdownToHtml(md) {
        return '<html><head><meta charset="utf-8"><style>body{font-family:Arial,sans-serif;max-width:800px;margin:2rem auto;padding:0 1rem;line-height:1.6}h1,h2,h3{color:#1e3a5f}table{border-collapse:collapse;width:100%;margin:1rem 0}td,th{border:1px solid #ddd;padding:8px;text-align:left}ul{padding-left:1.5rem}</style></head><body>' +
            renderMarkdown(md) +
            '</body></html>';
    }

    function downloadAsPdf(content) {
        var html = markdownToHtml(content);
        var iframe = document.createElement('iframe');
        iframe.style.position = 'fixed';
        iframe.style.right = '0';
        iframe.style.bottom = '0';
        iframe.style.width = '0';
        iframe.style.height = '0';
        iframe.style.border = 'none';
        document.body.appendChild(iframe);
        iframe.contentDocument.write(html);
        iframe.contentDocument.close();
        iframe.contentWindow.onafterprint = function () { document.body.removeChild(iframe); };
        setTimeout(function () { iframe.contentWindow.print(); }, 250);
    }

    function renderMarkdown(md) {
        var html = md
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');

        html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>');
        html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>');

        html = html.replace(/^---$/gm, '<hr>');

        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/g, '<a href="$2" target="_blank">$2</a>');

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
