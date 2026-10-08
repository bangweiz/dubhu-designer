// Run with mongosh while application writes are stopped. Requires a replica set.
// Preserves the legacy collection and all saved snapshots; safe to rerun.
const session = db.getMongo().startSession();
try {
  const database = session.getDatabase(db.getName());
  const drafts = db.draft_concierge_versions.find({});
  while (drafts.hasNext()) {
    const draft = drafts.next();
    session.withTransaction(() => {
      const filter = {_id: draft.concierge_id, organisation_id: draft.organisation_id};
      const concierge = database.concierges.findOne(filter);
      if (!concierge) throw new Error(`Missing parent for draft ${draft._id}`);
      if (concierge.next_version !== undefined) return;
      if (concierge.agents !== undefined) throw new Error(`Parent already has agents: ${concierge._id}`);
      if (!Array.isArray(draft.agents) || !Number.isInteger(draft.version) || draft.version < 1) {
        throw new Error(`Invalid draft ${draft._id}`);
      }
      const refs = (concierge.concierge_versions || []).filter(ref => !ref.concierge_version_id.equals(draft._id));
      database.concierges.updateOne(filter, {
        $set: {agents: draft.agents.map(agent => { const copy = {...agent}; delete copy.version; return copy; }), next_version: draft.version, concierge_versions: refs, updated_at: new Date(Math.max(Date.now(), concierge.updated_at.getTime() + 1))},
        $unset: {version: ""},
      });
    });
  }
  // Fail rather than silently expose an empty configuration for an unmigrated parent.
  if (db.concierges.countDocuments({next_version: {$exists: false}}) !== 0) {
    throw new Error('Some concierges have no draft to migrate; inspect them before starting the app.');
  }
  print('Concierge draft migration complete. Legacy drafts and saved snapshots were preserved.');
} finally {
  session.endSession();
}
