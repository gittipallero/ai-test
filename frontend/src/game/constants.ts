export const BLOCK_SIZE = 20;
export const ROWS = 21;
export const COLS = 19;

// 0: Empty, 1: Wall, 2: Dot, 3: Power, 9: Door
export const INITIAL_MAP = [
  [1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1],
  [1,2,2,2,2,2,2,2,2,1,2,2,2,2,2,2,2,2,1],
  [1,2,1,1,2,1,1,1,2,1,2,1,1,1,2,1,1,2,1],
  [1,2,1,1,2,1,1,1,2,1,2,1,1,1,2,1,1,2,1],
  [1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1],
  [1,2,1,1,2,1,2,1,1,1,1,1,2,1,2,1,1,2,1],
  [1,2,2,2,2,1,2,2,2,1,2,2,2,1,2,2,2,2,1],
  [1,1,1,1,2,1,1,1,0,1,0,1,1,1,2,1,1,1,1],
  [0,0,0,1,2,1,0,0,0,0,0,0,0,1,2,1,0,0,0],
  [1,1,1,1,2,1,0,1,1,9,1,1,0,1,2,1,1,1,1],
  [0,2,2,2,2,0,0,1,0,0,0,1,0,0,2,2,2,2,0], // teleport tunnel
  [1,1,1,1,2,1,0,1,1,1,1,1,0,1,2,1,1,1,1],
  [0,0,0,1,2,1,0,0,0,0,0,0,0,1,2,1,0,0,0],
  [1,1,1,1,2,1,2,1,1,1,1,1,2,1,2,1,1,1,1],
  [1,2,2,2,2,2,2,2,2,1,2,2,2,2,2,2,2,2,1],
  [1,2,1,1,2,1,1,1,2,1,2,1,1,1,2,1,1,2,1],
  [1,2,2,1,2,2,2,2,2,0,2,2,2,2,2,1,2,2,1],
  [1,1,2,1,2,1,2,1,1,1,1,1,2,1,2,1,2,1,1],
  [1,2,2,2,2,1,2,2,2,1,2,2,2,1,2,2,2,2,1],
  [1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1],
  [1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1],
];

export type Direction = 'UP' | 'DOWN' | 'LEFT' | 'RIGHT' | null;

export interface Position {
  x: number; // grid column
  y: number; // grid row
}

export interface Entity {
  pos: Position;
  dir: Direction;
  nextDir: Direction;
}

export interface GhostEntity {
    id: number;
    pos: Position;
    dir: Direction;
    color: string;
}

export const INITIAL_PACMAN: Position = { x: 9, y: 15 };

export const INITIAL_GHOSTS: GhostEntity[] = [
    { id: 1, pos: { x: 9, y: 7 }, dir: 'LEFT', color: 'red' },
    { id: 2, pos: { x: 9, y: 8 }, dir: 'RIGHT', color: 'pink' },
    { id: 3, pos: { x: 10, y: 7 }, dir: 'UP', color: 'cyan' },
    { id: 4, pos: { x: 10, y: 8 }, dir: 'DOWN', color: 'orange' },
];
