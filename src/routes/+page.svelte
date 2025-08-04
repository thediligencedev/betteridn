<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { ToggleGroup, ToggleGroupItem } from '$lib/components/ui/toggle-group';
	import { Badge } from '$lib/components/ui/badge';
	import { Card } from '$lib/components/ui/card';
	import { ArrowUp, ArrowDown, MessageSquare, TrendingUp, Clock, Plus } from 'lucide-svelte';
	
	type Post = {
		id: string;
		title: string;
		content: string;
		category: string;
		categoryColor: string;
		votes: number;
		comments: number;
		date: string;
		userVote?: 'up' | 'down' | null;
	};
	
	let sortBy = $state('top');
	
	let posts: Post[] = $state([
		{
			id: '1',
			title: "Citizens' Assembly for Climate Policy",
			content: "Establish a randomly selected citizens' assembly with the power to develop climate change policies. Members would be educated by experts across discip...",
			category: 'politics',
			categoryColor: 'bg-red-100 text-red-800',
			votes: 278,
			comments: 56,
			date: 'Jul 23',
			userVote: null
		},
		{
			id: '2',
			title: 'Implement Digital Voting System with Blockchain Verification',
			content: 'A secure digital voting system using blockchain technology would ensure transparency, reduce fraud, and increase accessibility. This could significant...',
			category: 'technology',
			categoryColor: 'bg-blue-100 text-blue-800',
			votes: 245,
			comments: 32,
			date: 'Jul 27',
			userVote: null
		},
		{
			id: '3',
			title: 'Mental Health First Responders Program',
			content: 'Create a specialized emergency response team of mental health professionals to respond to non-violent mental health crises instead of police. This wou...',
			category: 'healthcare',
			categoryColor: 'bg-green-100 text-green-800',
			votes: 219,
			comments: 28,
			date: 'Jul 12',
			userVote: null
		}
	]);
	
	function handleVote(postId: string, voteType: 'up' | 'down') {
		posts = posts.map(post => {
			if (post.id === postId) {
				const currentVote = post.userVote;
				let newVotes = post.votes;
				let newVote: 'up' | 'down' | null = null;
				
				if (currentVote === voteType) {
					// Remove vote
					newVotes = voteType === 'up' ? newVotes - 1 : newVotes + 1;
					newVote = null;
				} else if (currentVote === null) {
					// Add vote
					newVotes = voteType === 'up' ? newVotes + 1 : newVotes - 1;
					newVote = voteType;
				} else {
					// Change vote
					newVotes = voteType === 'up' ? newVotes + 2 : newVotes - 2;
					newVote = voteType;
				}
				
				return { ...post, votes: newVotes, userVote: newVote };
			}
			return post;
		});
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<ToggleGroup bind:value={sortBy} type="single">
			<ToggleGroupItem value="top" class="gap-2">
				<TrendingUp class="h-4 w-4" />
				Top
			</ToggleGroupItem>
			<ToggleGroupItem value="new" class="gap-2">
				<Clock class="h-4 w-4" />
				New
			</ToggleGroupItem>
			<ToggleGroupItem value="trending" class="gap-2">
				<TrendingUp class="h-4 w-4" />
				Trending
			</ToggleGroupItem>
		</ToggleGroup>
		
		<Button class="gap-2">
			<Plus class="h-4 w-4" />
			Create New Post
		</Button>
	</div>
	
	<div class="space-y-4">
		{#each posts as post (post.id)}
			<Card class="p-4">
				<div class="flex gap-4">
					<div class="flex flex-col items-center">
						<button
							onclick={() => handleVote(post.id, 'up')}
							class="p-1 rounded hover:bg-gray-100 transition-colors"
							class:text-orange-500={post.userVote === 'up'}
						>
							<ArrowUp class="h-5 w-5" />
						</button>
						<span class="text-lg font-semibold my-1">{post.votes}</span>
						<button
							onclick={() => handleVote(post.id, 'down')}
							class="p-1 rounded hover:bg-gray-100 transition-colors"
							class:text-blue-500={post.userVote === 'down'}
						>
							<ArrowDown class="h-5 w-5" />
						</button>
					</div>
					
					<div class="flex-1">
						<Badge variant="secondary" class={post.categoryColor + ' mb-2'}>
							{post.category}
						</Badge>
						
						<h3 class="text-xl font-semibold mb-2">{post.title}</h3>
						
						<p class="text-gray-600 mb-3">{post.content}</p>
						
						<div class="flex items-center gap-4 text-sm text-gray-500">
							<button class="flex items-center gap-1 hover:text-gray-700">
								<MessageSquare class="h-4 w-4" />
								{post.comments} replies
							</button>
							<span>{post.date}</span>
						</div>
					</div>
				</div>
			</Card>
		{/each}
	</div>
</div>