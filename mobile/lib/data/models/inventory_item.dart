class InventoryItem {
  final int id;
  final String name;
  final int quantity;
  final double weight;
  final String category;
  final String damageDice;
  final String damageType;
  final String weaponProperties;
  final int acBonus;
  final String armorType;
  final String description;
  final bool isEquipped;
  final bool isMagical;
  final bool attunement;
  final bool isIdentified;
  final String notes;
  final String attackAbility;
  final int attackBonus;
  const InventoryItem({
    required this.id,
    required this.name,
    this.quantity = 1,
    this.weight = 0,
    this.category = '',
    this.damageDice = '',
    this.damageType = '',
    this.weaponProperties = '',
    this.acBonus = 0,
    this.armorType = '',
    this.description = '',
    this.isEquipped = false,
    this.isMagical = false,
    this.attunement = false,
    this.isIdentified = true,
    this.notes = '',
    this.attackAbility = '',
    this.attackBonus = 0,
  });
  factory InventoryItem.fromJson(Map<String, dynamic> j) => InventoryItem(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    quantity: (j['quantity'] as num?)?.toInt() ?? 1,
    weight: (j['weight'] as num?)?.toDouble() ?? 0,
    category: j['category'] as String? ?? '',
    damageDice: j['damage_dice'] as String? ?? j['damageDice'] as String? ?? '',
    damageType: j['damage_type'] as String? ?? j['damageType'] as String? ?? '',
    weaponProperties:
        j['weapon_properties'] as String? ??
        j['weaponProperties'] as String? ??
        '',
    acBonus:
        (j['ac_bonus'] as num?)?.toInt() ??
        (j['acBonus'] as num?)?.toInt() ??
        0,
    armorType: j['armor_type'] as String? ?? j['armorType'] as String? ?? '',
    description: j['description'] as String? ?? '',
    isEquipped: j['is_equipped'] as bool? ?? j['isEquipped'] as bool? ?? false,
    isMagical: j['is_magical'] as bool? ?? j['isMagical'] as bool? ?? false,
    attunement: j['attunement'] as bool? ?? false,
    isIdentified:
        j['is_identified'] as bool? ?? j['isIdentified'] as bool? ?? true,
    notes: j['notes'] as String? ?? '',
    attackAbility:
        j['attack_ability'] as String? ?? j['attackAbility'] as String? ?? '',
    attackBonus:
        (j['attack_bonus'] as num?)?.toInt() ??
        (j['attackBonus'] as num?)?.toInt() ??
        0,
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'quantity': quantity,
    'weight': weight,
    'category': category,
    'damage_dice': damageDice,
    'damage_type': damageType,
    'weapon_properties': weaponProperties,
    'ac_bonus': acBonus,
    'armor_type': armorType,
    'description': description,
    'is_equipped': isEquipped,
    'is_magical': isMagical,
    'attunement': attunement,
    'is_identified': isIdentified,
    'notes': notes,
    'attack_ability': attackAbility,
    'attack_bonus': attackBonus,
  };
  @override
  bool operator ==(Object other) =>
      other is InventoryItem &&
      other.id == id &&
      other.name == name &&
      other.quantity == quantity &&
      other.weight == weight &&
      other.category == category &&
      other.damageDice == damageDice &&
      other.damageType == damageType &&
      other.weaponProperties == weaponProperties &&
      other.acBonus == acBonus &&
      other.armorType == armorType &&
      other.description == description &&
      other.isEquipped == isEquipped &&
      other.isMagical == isMagical &&
      other.attunement == attunement &&
      other.isIdentified == isIdentified &&
      other.notes == notes &&
      other.attackAbility == attackAbility &&
      other.attackBonus == attackBonus;
  @override
  int get hashCode => Object.hashAll([
    id,
    name,
    quantity,
    weight,
    category,
    damageDice,
    damageType,
    weaponProperties,
    acBonus,
    armorType,
    description,
    isEquipped,
    isMagical,
    attunement,
    isIdentified,
    notes,
    attackAbility,
    attackBonus,
  ]);
}
