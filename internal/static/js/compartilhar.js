(function () {
    var spinner = document.getElementById('spinner');
    var errorMsg = document.getElementById('error-message');
    var contentSection = document.getElementById('content-section');
    var contentTitle = document.getElementById('content-title');
    var metadata = document.getElementById('metadata');
    var contentBody = document.getElementById('content-body');
    var downloadMd = document.getElementById('download-md');
    var downloadPdf = document.getElementById('download-pdf');

    var currentContent = '';
    var currentId = '';
    var currentType = '';

    var params = new URLSearchParams(window.location.search);
    var resumoId = params.get('resumo');
    var analiseId = params.get('analise');

    if (resumoId) {
        currentId = resumoId;
        currentType = 'resumo';
        document.title = 'Resumo - smartLattes';
        loadContent('/api/summary/view/' + resumoId, 'Resumo do Pesquisador');
    } else if (analiseId) {
        currentId = analiseId;
        currentType = 'analise';
        document.title = 'An\u00e1lise de Rela\u00e7\u00f5es - smartLattes';
        loadContent('/api/analysis/view/' + analiseId, 'An\u00e1lise de Rela\u00e7\u00f5es');
    } else {
        showError('Link inv\u00e1lido. Nenhum resumo ou an\u00e1lise especificado.');
    }

    function loadContent(url, title) {
        spinner.classList.add('visible');

        fetch(url)
            .then(function (r) {
                return r.json().then(function (data) {
                    return { status: r.status, body: data };
                });
            })
            .then(function (result) {
                spinner.classList.remove('visible');

                if (result.status === 404) {
                    showError(result.body.error || 'Conte\u00fado n\u00e3o encontrado.');
                    return;
                }

                if (!result.body.success) {
                    showError(result.body.error || 'Erro ao carregar conte\u00fado.');
                    return;
                }

                var text = result.body.summary || result.body.analysis || '';
                currentContent = text;

                contentTitle.textContent = title;

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

                contentBody.innerHTML = renderMarkdown(text);
                contentSection.style.display = 'block';
            })
            .catch(function () {
                spinner.classList.remove('visible');
                showError('Erro de conex\u00e3o ao carregar conte\u00fado.');
            });
    }

    downloadMd.addEventListener('click', function () {
        var prefix = currentType === 'resumo' ? 'resumo-' : 'analise-';
        downloadBlob(currentContent, prefix + currentId + '.md', 'text/markdown');
    });
    downloadPdf.addEventListener('click', function () {
        downloadAsPdf(currentContent);
    });

    function showError(message) {
        errorMsg.textContent = message;
        errorMsg.classList.add('visible');
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
