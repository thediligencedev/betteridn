<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	
	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>BetterIDN - Ideas to improve Indonesia</title>
</svelte:head>

<div class="min-h-screen bg-background">
	<header class="border-b">
		<div class="container mx-auto px-4 py-4">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-2xl font-bold">BetterIDN</h1>
					<p class="text-sm text-muted-foreground">Ideas to improve governance, country, and society</p>
				</div>
			</div>
		</div>
	</header>
	
	<main class="container mx-auto px-4 py-8">
		{@render children?.()}
	</main>
</div>
