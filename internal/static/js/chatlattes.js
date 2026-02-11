(function () {
    var settingsPanel = document.getElementById('chat-settings');
    var chatArea = document.getElementById('chat-area');
    var providerSelect = document.getElementById('provider-select');
    var apiKeyInput = document.getElementById('api-key-input');
    var loadModelsBtn = document.getElementById('load-models-btn');
    var modelSelect = document.getElementById('model-select');
    var startChatBtn = document.getElementById('start-chat-btn');
    var settingsError = document.getElementById('settings-error');
    var chatMessages = document.getElementById('chat-messages');
    var chatInput = document.getElementById('chat-input');
    var sendBtn = document.getElementById('send-btn');
    var newChatBtn = document.getElementById('new-chat-btn');
    var providerInfo = document.getElementById('chat-provider-info');

    var messages = [];
    var isWaiting = false;

    providerSelect.addEventListener('change', checkLoadModels);
    apiKeyInput.addEventListener('input', checkLoadModels);

    function checkLoadModels() {
        loadModelsBtn.disabled = !(providerSelect.value && apiKeyInput.value.length >= 10);
    }

    loadModelsBtn.addEventListener('click', function () {
        hideSettingsError();
        loadModelsBtn.disabled = true;
        loadModelsBtn.textContent = 'Carregando...';

        fetch('/api/models', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                provider: providerSelect.value,
                apiKey: apiKeyInput.value
            })
        })
        .then(function (r) {
            return r.json().then(function (data) {
                return { status: r.status, body: data };
            });
        })
        .then(function (result) {
            loadModelsBtn.disabled = false;
            loadModelsBtn.textContent = 'Carregar Modelos';

            if (!result.body.success) {
                showSettingsError(result.body.error || 'Erro ao carregar modelos');
                return;
            }

            modelSelect.innerHTML = '<option value="">Selecione o modelo...</option>';
            for (var i = 0; i < result.body.models.length; i++) {
                var opt = document.createElement('option');
                opt.value = result.body.models[i].id;
                opt.textContent = result.body.models[i].displayName || result.body.models[i].id;
                modelSelect.appendChild(opt);
            }
            modelSelect.disabled = false;
        })
        .catch(function () {
            loadModelsBtn.disabled = false;
            loadModelsBtn.textContent = 'Carregar Modelos';
            showSettingsError('Erro de conexão ao carregar modelos');
        });
    });

    modelSelect.addEventListener('change', function () {
        startChatBtn.disabled = !modelSelect.value;
    });

    startChatBtn.addEventListener('click', function () {
        settingsPanel.style.display = 'none';
        chatArea.style.display = 'flex';
        var providerName = providerSelect.options[providerSelect.selectedIndex].text;
        var modelName = modelSelect.options[modelSelect.selectedIndex].text;
        providerInfo.textContent = providerName + ' / ' + modelName;
        chatInput.focus();
        updateSendBtn();
    });

    newChatBtn.addEventListener('click', function () {
        messages = [];
        chatMessages.innerHTML = '';
        showWelcome();
        chatInput.focus();
    });

    chatInput.addEventListener('input', function () {
        this.style.height = 'auto';
        this.style.height = Math.min(this.scrollHeight, 150) + 'px';
        updateSendBtn();
    });

    chatInput.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });

    sendBtn.addEventListener('click', function () {
        sendMessage();
    });

    var suggestionBtns = document.querySelectorAll('.suggestion-btn');
    for (var i = 0; i < suggestionBtns.length; i++) {
        suggestionBtns[i].addEventListener('click', function () {
            chatInput.value = this.textContent;
            sendMessage();
        });
    }

    function sendMessage() {
        var text = chatInput.value.trim();
        if (!text || isWaiting) return;

        removeWelcome();
        removeError();

        messages.push({ role: 'user', content: text });
        appendMessage('user', text);

        chatInput.value = '';
        chatInput.style.height = 'auto';
        updateSendBtn();

        isWaiting = true;
        sendBtn.disabled = true;
        showTyping();
        scrollToBottom();

        fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                provider: providerSelect.value,
                apiKey: apiKeyInput.value,
                model: modelSelect.value,
                messages: messages
            })
        })
        .then(function (r) {
            return r.json().then(function (data) {
                return { status: r.status, body: data };
            });
        })
        .then(function (result) {
            removeTyping();
            isWaiting = false;
            updateSendBtn();

            if (!result.body.success) {
                showChatError(result.body.error || 'Erro ao obter resposta');
                messages.pop();
                return;
            }

            messages.push({ role: 'assistant', content: result.body.response });
            appendMessage('assistant', result.body.response);
            scrollToBottom();
        })
        .catch(function () {
            removeTyping();
            isWaiting = false;
            updateSendBtn();
            showChatError('Erro de conexão. Verifique sua rede e tente novamente.');
            messages.pop();
        });
    }

    function appendMessage(role, content) {
        var div = document.createElement('div');
        div.className = 'chat-message chat-message-' + role;

        var avatar = document.createElement('div');
        avatar.className = 'chat-avatar';
        avatar.textContent = role === 'user' ? '\uD83D\uDC64' : '\uD83E\uDD16';

        var bubble = document.createElement('div');
        bubble.className = 'chat-bubble';

        if (role === 'assistant') {
            bubble.innerHTML = renderMarkdown(content);
        } else {
            bubble.textContent = content;
        }

        div.appendChild(avatar);
        div.appendChild(bubble);
        chatMessages.appendChild(div);
    }

    function showTyping() {
        var div = document.createElement('div');
        div.className = 'chat-typing';
        div.id = 'typing-indicator';

        var avatar = document.createElement('div');
        avatar.className = 'chat-avatar';
        avatar.style.backgroundColor = 'var(--color-nav-bg)';
        avatar.style.color = 'white';
        avatar.textContent = '\uD83E\uDD16';

        var dots = document.createElement('div');
        dots.className = 'typing-dots';
        dots.innerHTML = '<span></span><span></span><span></span>';

        div.appendChild(avatar);
        div.appendChild(dots);
        chatMessages.appendChild(div);
    }

    function removeTyping() {
        var el = document.getElementById('typing-indicator');
        if (el) el.remove();
    }

    function showWelcome() {
        var welcome = document.createElement('div');
        welcome.className = 'chat-welcome';
        welcome.id = 'chat-welcome';
        welcome.innerHTML =
            '<div class="chat-welcome-icon">&#128172;</div>' +
            '<h3>Bem-vindo ao chatLattes</h3>' +
            '<p>Fa\u00e7a perguntas sobre os curr\u00edculos Lattes armazenados na base de dados.</p>' +
            '<div class="chat-suggestions">' +
            '<p class="suggestions-label">Experimente perguntar:</p>' +
            '<button class="suggestion-btn">Quantos pesquisadores na base?</button>' +
            '<button class="suggestion-btn">\u00c1reas de atua\u00e7\u00e3o mais comuns</button>' +
            '<button class="suggestion-btn">Pesquisadores e especialidades</button>' +
            '</div>';
        chatMessages.appendChild(welcome);

        var btns = welcome.querySelectorAll('.suggestion-btn');
        for (var i = 0; i < btns.length; i++) {
            btns[i].addEventListener('click', function () {
                chatInput.value = this.textContent;
                sendMessage();
            });
        }
    }

    function removeWelcome() {
        var el = document.getElementById('chat-welcome');
        if (el) el.remove();
        var existing = chatMessages.querySelector('.chat-welcome');
        if (existing) existing.remove();
    }

    function showChatError(msg) {
        removeError();
        var div = document.createElement('div');
        div.className = 'chat-error';
        div.id = 'chat-error';
        div.textContent = msg;
        chatMessages.appendChild(div);
        scrollToBottom();
    }

    function removeError() {
        var el = document.getElementById('chat-error');
        if (el) el.remove();
    }

    function showSettingsError(msg) {
        settingsError.textContent = msg;
        settingsError.classList.add('visible');
    }

    function hideSettingsError() {
        settingsError.classList.remove('visible');
    }

    function updateSendBtn() {
        sendBtn.disabled = !chatInput.value.trim() || isWaiting;
    }

    function scrollToBottom() {
        chatMessages.scrollTop = chatMessages.scrollHeight;
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

        html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/g, '<a href="$2" target="_blank">$1</a>');

        html = html.replace(/^\|(.+)\|$/gm, function (match, content) {
            var cells = content.split('|').map(function (c) { return c.trim(); });
            return '<tr>' + cells.map(function (c) {
                if (/^[-:]+$/.test(c)) return '';
                return '<td>' + c + '</td>';
            }).join('') + '</tr>';
        });
        html = html.replace(/((?:<tr>.*?<\/tr>\n?)+)/g, '<table>$1</table>');
        html = html.replace(/<tr><\/tr>/g, '');

        html = html.replace(/^\d+\.\s(.+)$/gm, '<oli>$1</oli>');
        html = html.replace(/((?:<oli>.*?<\/oli>\n?)+)/g, function (match) {
            var items = match.replace(/<\/?oli>/g, function (tag) {
                return tag === '<oli>' ? '<li>' : '</li>';
            });
            return '<ol>' + items + '</ol>';
        });

        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');
        html = html.replace(/((?:<li>.*?<\/li>\n?)+)/g, '<ul>$1</ul>');

        html = html.replace(/^(?!<[hultdo]|\s*$)(.+)$/gm, '<p>$1</p>');
        html = html.replace(/<p>\s*<\/p>/g, '');

        return html;
    }
})();
