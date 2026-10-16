class Currency {
  final int cp, sp, ep, gp, pp;
  const Currency({
    this.cp = 0,
    this.sp = 0,
    this.ep = 0,
    this.gp = 0,
    this.pp = 0,
  });
  factory Currency.fromJson(Map<String, dynamic> j) => Currency(
    cp: (j['cp'] as num?)?.toInt() ?? 0,
    sp: (j['sp'] as num?)?.toInt() ?? 0,
    ep: (j['ep'] as num?)?.toInt() ?? 0,
    gp: (j['gp'] as num?)?.toInt() ?? 0,
    pp: (j['pp'] as num?)?.toInt() ?? 0,
  );
  Map<String, dynamic> toJson() => {
    'cp': cp,
    'sp': sp,
    'ep': ep,
    'gp': gp,
    'pp': pp,
  };
  @override
  bool operator ==(Object other) =>
      other is Currency &&
      other.cp == cp &&
      other.sp == sp &&
      other.ep == ep &&
      other.gp == gp &&
      other.pp == pp;
  @override
  int get hashCode => Object.hash(cp, sp, ep, gp, pp);
}
