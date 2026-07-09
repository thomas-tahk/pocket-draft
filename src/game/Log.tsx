export function Log({ lines }: { lines: string[] }) {
  return (
    <div style={{ maxHeight: 160, overflowY: 'auto', fontSize: 12, lineHeight: 1.5, border: '1px solid #33363d', borderRadius: 8, padding: 8 }}>
      {lines.length === 0 ? <div style={{ opacity: 0.6 }}>No events yet.</div> : lines.map((l, i) => <div key={i}>{l}</div>)}
    </div>
  );
}
