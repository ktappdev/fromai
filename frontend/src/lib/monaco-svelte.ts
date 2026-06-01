export function registerSvelteLanguage(monaco: any) {
	monaco.languages.register({ id: 'svelte' });
	monaco.languages.setLanguageConfiguration('svelte', {
		comments: { lineComment: '//', blockComment: ['<!--', '-->'] },
		brackets: [['{', '}'], ['[', ']'], ['(', ')'], ['<', '>']],
		autoClosingPairs: [
			{ open: '{', close: '}' },
			{ open: '[', close: ']' },
			{ open: '(', close: ')' },
			{ open: '"', close: '"' },
			{ open: "'", close: "'" },
			{ open: '<', close: '>' }
		],
		surroundingPairs: [
			{ open: '{', close: '}' },
			{ open: '[', close: ']' },
			{ open: '(', close: ')' },
			{ open: '"', close: '"' },
			{ open: "'", close: "'" },
			{ open: '<', close: '>' }
		]
	});
	monaco.languages.setMonarchTokensProvider('svelte', {
		keywords: ['if', 'else', 'each', 'await', 'then', 'catch', 'key', 'slot', 'let', 'html', 'debug'],
		tokenizer: {
			root: [
				[/\{#([a-z]+)/, { token: 'tag', next: '@block.$1' }],
				[/\{\//, { token: 'tag', next: '@endblock' }],
				[/\{/, 'delimiter.bracket'],
				[/\}/, 'delimiter.bracket'],
				[/<!--/, 'comment', '@comment'],
				[/<script/, { token: 'tag', next: '@script' }],
				[/<style/, { token: 'tag', next: '@style' }],
				[/<[a-zA-Z][a-zA-Z0-9]*/, 'tag'],
				[/<\/[a-zA-Z][a-zA-Z0-9]*/, 'tag'],
				[/&[a-zA-Z]+;/, 'string.escape'],
				[/"[^"]*"/, 'string'],
				[/'[^']*'/, 'string'],
				[/\b(true|false|null|undefined)\b/, 'constant.language'],
				[/\b(\d+)\b/, 'number'],
				[/\b([a-zA-Z_$][a-zA-Z0-9_$]*)\b/, 'identifier']
			],
			block: [
				[/}/, { token: 'tag', next: '@pop' }],
				[/[a-zA-Z_$][a-zA-Z0-9_$]*/, 'variable'],
				[/\./, 'delimiter'],
				[/"[^"]*"/, 'string'],
				[/'[^']*'/, 'string'],
				[/\d+/, 'number'],
				[/\S+/, 'variable']
			],
			endblock: [[/}/, { token: 'tag', next: '@pop' }]],
			comment: [[/-->/, 'comment', '@pop'], [/[^-]+/, 'comment'], [/-/, 'comment']],
			script: [
				[/<\/script>/, { token: 'tag', next: '@pop' }],
				[/\b(import|export|from|const|let|var|function|return|if|else|for|while|class|extends|new|this|async|await)\b/, 'keyword'],
				[/\b(true|false|null|undefined)\b/, 'constant.language'],
				[/"[^"]*"/, 'string'],
				[/'[^']*'/, 'string'],
				[/\d+/, 'number'],
				[/\/\*/, 'comment', '@scriptComment'],
				[/\/\//, 'comment'],
				[/\S+/, 'identifier']
			],
			scriptComment: [[/\*\//, 'comment', '@pop'], [/[^*]+/, 'comment'], [/\*/, 'comment']],
			style: [
				[/<\/style>/, { token: 'tag', next: '@pop' }],
				[/[.#][a-zA-Z_-][a-zA-Z0-9_-]*/, 'tag.selector'],
				[/[a-zA-Z-]+(?=:)/, 'tag.property'],
				[/:[a-zA-Z-]+/, 'tag.value'],
				[/\{/, 'delimiter.bracket'],
				[/\}/, 'delimiter.bracket'],
				[/\/\*/, 'comment', '@styleComment'],
				[/[^\{\}]+/, 'identifier']
			],
			styleComment: [[/\*\//, 'comment', '@pop'], [/[^*]+/, 'comment'], [/\*/, 'comment']]
		}
	});
}