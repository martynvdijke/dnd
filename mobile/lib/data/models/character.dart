import 'currency.dart';
import 'spellcasting.dart';
import 'spell.dart';
import 'inventory_item.dart';
import 'feature.dart';
import 'char_class.dart';

int _mod(int s) => ((s - 10) / 2).floor();

class Character {
  final int id;
  final int userId;
  final String name;
  final String race;
  final String class_;
  final String subclass;
  final int level;
  final int xp;
  final String background;
  final String alignment;
  final int str;
  final int dex;
  final int con;
  final int int_;
  final int wis;
  final int cha;
  final int ac;
  final int initiative;
  final int speed;
  final int hpMax;
  final int hpCurrent;
  final int tempHp;
  final String hitDice;
  final int hitDiceCurrent;
  final int proficiencyBonus;
  final bool inspiration;
  final int passivePerception;
  final int deathSavesSuccesses;
  final int deathSavesFailures;
  final String concentratingOn;
  final String portraitUrl;
  final String characterType;
  final bool canEdit;
  final List<Feature> features;
  final Spellcasting? spellcasting;
  final List<Spell> spells;
  final List<InventoryItem> inventory;
  final Currency? currency;
  final List<CharClass> classes;
  final List<dynamic> campaigns;
  const Character({
    required this.id,
    required this.userId,
    required this.name,
    required this.race,
    required this.class_,
    required this.subclass,
    required this.level,
    required this.xp,
    required this.background,
    required this.alignment,
    required this.str,
    required this.dex,
    required this.con,
    required this.int_,
    required this.wis,
    required this.cha,
    required this.ac,
    required this.initiative,
    required this.speed,
    required this.hpMax,
    required this.hpCurrent,
    required this.tempHp,
    required this.hitDice,
    required this.hitDiceCurrent,
    required this.proficiencyBonus,
    required this.inspiration,
    required this.passivePerception,
    required this.deathSavesSuccesses,
    required this.deathSavesFailures,
    required this.concentratingOn,
    required this.portraitUrl,
    required this.characterType,
    required this.canEdit,
    this.features = const [],
    this.spellcasting,
    this.spells = const [],
    this.inventory = const [],
    this.currency,
    this.classes = const [],
    this.campaigns = const [],
  });
  int get strMod => _mod(str);
  int get dexMod => _mod(dex);
  int get conMod => _mod(con);
  int get intMod => _mod(int_);
  int get wisMod => _mod(wis);
  int get chaMod => _mod(cha);
  int get spellSaveDc => spellcasting?.saveDc ?? 0;
  int get spellAttackBonus => spellcasting?.attackBonus ?? 0;
  factory Character.fromJson(Map<String, dynamic> j) => Character(
    id: (j['id'] as num).toInt(),
    userId: (j['user_id'] as num?)?.toInt() ?? 0,
    name: j['name'] as String? ?? '',
    race: j['race'] as String? ?? '',
    class_: j['class'] as String? ?? '',
    subclass: j['subclass'] as String? ?? '',
    level: (j['level'] as num?)?.toInt() ?? 1,
    xp: (j['xp'] as num?)?.toInt() ?? 0,
    background: j['background'] as String? ?? '',
    alignment: j['alignment'] as String? ?? '',
    str: (j['str'] as num?)?.toInt() ?? 10,
    dex: (j['dex'] as num?)?.toInt() ?? 10,
    con: (j['con'] as num?)?.toInt() ?? 10,
    int_: (j['int'] as num?)?.toInt() ?? 10,
    wis: (j['wis'] as num?)?.toInt() ?? 10,
    cha: (j['cha'] as num?)?.toInt() ?? 10,
    ac: (j['ac'] as num?)?.toInt() ?? 10,
    initiative: (j['initiative'] as num?)?.toInt() ?? 0,
    speed: (j['speed'] as num?)?.toInt() ?? 30,
    hpMax: (j['hp_max'] as num?)?.toInt() ?? 0,
    hpCurrent: (j['hp_current'] as num?)?.toInt() ?? 0,
    tempHp: (j['temp_hp'] as num?)?.toInt() ?? 0,
    hitDice: j['hit_dice'] as String? ?? '',
    hitDiceCurrent: (j['hit_dice_current'] as num?)?.toInt() ?? 0,
    proficiencyBonus: (j['proficiency_bonus'] as num?)?.toInt() ?? 2,
    inspiration: j['inspiration'] as bool? ?? false,
    passivePerception: (j['passive_perception'] as num?)?.toInt() ?? 10,
    deathSavesSuccesses: (j['death_saves_successes'] as num?)?.toInt() ?? 0,
    deathSavesFailures: (j['death_saves_failures'] as num?)?.toInt() ?? 0,
    concentratingOn: j['concentrating_on'] as String? ?? '',
    portraitUrl: j['portrait_url'] as String? ?? '',
    characterType: j['character_type'] as String? ?? '',
    canEdit: j['can_edit'] as bool? ?? false,
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'user_id': userId,
    'name': name,
    'race': race,
    'class': class_,
    'subclass': subclass,
    'level': level,
    'xp': xp,
    'background': background,
    'alignment': alignment,
    'str': str,
    'dex': dex,
    'con': con,
    'int': int_,
    'wis': wis,
    'cha': cha,
    'ac': ac,
    'initiative': initiative,
    'speed': speed,
    'hp_max': hpMax,
    'hp_current': hpCurrent,
    'temp_hp': tempHp,
    'hit_dice': hitDice,
    'hit_dice_current': hitDiceCurrent,
    'proficiency_bonus': proficiencyBonus,
    'inspiration': inspiration,
    'passive_perception': passivePerception,
    'death_saves_successes': deathSavesSuccesses,
    'death_saves_failures': deathSavesFailures,
    'concentrating_on': concentratingOn,
    'portrait_url': portraitUrl,
    'character_type': characterType,
    'can_edit': canEdit,
  };
  Character copyWith({
    int? hpCurrent,
    int? tempHp,
    Currency? currency,
    Spellcasting? spellcasting,
  }) => Character(
    id: id,
    userId: userId,
    name: name,
    race: race,
    class_: class_,
    subclass: subclass,
    level: level,
    xp: xp,
    background: background,
    alignment: alignment,
    str: str,
    dex: dex,
    con: con,
    int_: int_,
    wis: wis,
    cha: cha,
    ac: ac,
    initiative: initiative,
    speed: speed,
    hpMax: hpMax,
    hpCurrent: hpCurrent ?? this.hpCurrent,
    tempHp: tempHp ?? this.tempHp,
    hitDice: hitDice,
    hitDiceCurrent: hitDiceCurrent,
    proficiencyBonus: proficiencyBonus,
    inspiration: inspiration,
    passivePerception: passivePerception,
    deathSavesSuccesses: deathSavesSuccesses,
    deathSavesFailures: deathSavesFailures,
    concentratingOn: concentratingOn,
    portraitUrl: portraitUrl,
    characterType: characterType,
    canEdit: canEdit,
    features: features,
    spellcasting: spellcasting ?? this.spellcasting,
    spells: spells,
    inventory: inventory,
    currency: currency ?? this.currency,
    classes: classes,
    campaigns: campaigns,
  );
  @override
  bool operator ==(Object other) =>
      other is Character &&
      other.id == id &&
      other.userId == userId &&
      other.name == name &&
      other.race == race &&
      other.class_ == class_ &&
      other.subclass == subclass &&
      other.level == level &&
      other.xp == xp &&
      other.background == background &&
      other.alignment == alignment &&
      other.str == str &&
      other.dex == dex &&
      other.con == con &&
      other.int_ == int_ &&
      other.wis == wis &&
      other.cha == cha &&
      other.ac == ac &&
      other.initiative == initiative &&
      other.speed == speed &&
      other.hpMax == hpMax &&
      other.hpCurrent == hpCurrent &&
      other.tempHp == tempHp &&
      other.hitDice == hitDice &&
      other.hitDiceCurrent == hitDiceCurrent &&
      other.proficiencyBonus == proficiencyBonus &&
      other.inspiration == inspiration &&
      other.passivePerception == passivePerception &&
      other.deathSavesSuccesses == deathSavesSuccesses &&
      other.deathSavesFailures == deathSavesFailures &&
      other.concentratingOn == concentratingOn &&
      other.portraitUrl == portraitUrl &&
      other.characterType == characterType &&
      other.canEdit == canEdit;
  @override
  int get hashCode => Object.hashAll([
    id,
    userId,
    name,
    race,
    class_,
    subclass,
    level,
    xp,
    background,
    alignment,
    str,
    dex,
    con,
    int_,
    wis,
    cha,
    ac,
    initiative,
    speed,
    hpMax,
    hpCurrent,
    tempHp,
    hitDice,
    hitDiceCurrent,
    proficiencyBonus,
    inspiration,
    passivePerception,
    deathSavesSuccesses,
    deathSavesFailures,
    concentratingOn,
    portraitUrl,
    characterType,
    canEdit,
  ]);
}
