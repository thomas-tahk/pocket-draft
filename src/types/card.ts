export type EnergyType =
  | 'Grass'
  | 'Fire'
  | 'Water'
  | 'Lightning'
  | 'Psychic'
  | 'Fighting'
  | 'Darkness'
  | 'Metal'
  | 'Dragon'
  | 'Colorless';

export type CardType = EnergyType | 'Trainer';

export type TrainerKind = 'Item' | 'Supporter';

export type Stage = 'Basic' | 'Stage 1' | 'Stage 2';

export type Rarity =
  | '◊'
  | '◊◊'
  | '◊◊◊'
  | '◊◊◊◊'
  | '☆'
  | '☆☆'
  | '☆☆☆'
  | '♕'
  | 'Promo';

export type ExKind = 'regular' | 'ex' | 'mega-ex';

export type Attack = {
  cost: string;
  name: string;
  damage: string;
  effect: string;
};

export type Ability = {
  name: string;
  effect: string;
};

export type RawCard = {
  id: string;
  setId: string;
  number: number;
  name: string;
  rarity: Rarity;
  exKind: ExKind;
  cardType: CardType;
  trainerKind: TrainerKind | null;
  hp: number | null;
  stage: Stage | null;
  evolvesFrom: string | null;
  attacks: Attack[];
  ability: Ability | null;
  weakness: EnergyType | null;
  retreat: number | null;
  pack: string;
  trainerText: string | null;
  imageThumb: string;
  imageFull: string;
  artist: string | null;
  flavor: string | null;
};

export type SetInfo = {
  id: string;
  name: string;
  packs: string[];
  cardCount: number;
};

export type DraftableRarity = '◊' | '◊◊' | '◊◊◊' | '◊◊◊◊';

export type Card = RawCard & {
  draftRarity: DraftableRarity;
  isEx: boolean;
  isMegaEx: boolean;
};
