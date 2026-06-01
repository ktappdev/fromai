export interface BadgeDef {
	id: string;
	name: string;
	description: string;
	icon: string;
}

export const BADGE_CATALOG: BadgeDef[] = [
	{ id: 'first_task', name: 'First Steps', description: 'Complete your first task', icon: '🎯' },
	{ id: 'streak_3', name: 'On Fire', description: 'Maintain a 3-day streak', icon: '🔥' },
	{ id: 'streak_7', name: 'Week Warrior', description: 'Maintain a 7-day streak', icon: '⚔️' },
	{ id: 'streak_30', name: 'Monthly Master', description: 'Maintain a 30-day streak', icon: '🏆' },
	{ id: 'perfect_five', name: 'Perfect Five', description: 'Get 5 A grades in a row', icon: '⭐' },
	{ id: 'speed_runner', name: 'Speed Runner', description: 'Complete 5 tasks in under 10 minutes each', icon: '⚡' },
	{ id: 'polyglot', name: 'Polyglot', description: 'Complete tasks in 5 different languages', icon: '🌍' },
	{ id: 'algorithm_master', name: 'Algorithm Master', description: 'Complete 10 algorithm category tasks', icon: '🧮' },
	{ id: 'daily_champion', name: 'Daily Champion', description: 'Complete 7 daily challenges', icon: '👑' },
	{ id: 'century', name: 'Century', description: 'Complete 100 total tasks', icon: '💯' },
];

export function getBadgeDef(id: string): BadgeDef | undefined {
	return BADGE_CATALOG.find((b) => b.id === id);
}