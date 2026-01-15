# nbq

The notebook tool for the Coding Agent era—built so agents can reason about Jupyter files without begging a GUI.
nbq is a tiny CLI that treats notebooks as structured documents, not blobs of JSON.
Its core move is slicing Markdown/code cell pairs so the surrounding story travels with the code.
It queries headings, parameter blocks, code cells, and outputs straight from the shell.
Massive JSON notebooks are painful to diff or skim, so nbq exposes the slices directly.
Think `rg` + `jq`, but tuned for notebook semantics instead of raw JSON.
Ships as a single Go binary so you can drop it into any CI or test runner.

## Commands (stdin → stdout)

All commands read an `.ipynb` JSON from stdin (or `--file`) and write the result to stdout so they can be piped like `jq`.

- `nbq toc` – emit Markdown headings with a short preview (`--words`, `--format md|json`). JSON returns heading-only cells annotated with `_index` (original cell position).
- `nbq section` – start from a query (see below) and return Markdown+code sets, including nested subsections until the next peer heading (`--sets`, `--format md|json|py`). JSON returns the selected cells with `_index` fields.
- `nbq cells` – same as `section` but anchored by absolute cell order (`--query start:40`). JSON also emits `_index`-annotated cells.
- `nbq outputs` – extract the outputs array for the matched cell (`--outputs text|mime=image/png|raw`).

### Queries (shared by section/cells/outputs)

Use `--query TYPE:VALUE` to locate the starting cell. Multiple `--query` flags are ANDed.

- `start:37` – absolute cell index.
- `match:"Crossref"` – substring or regex against headings/Markdown/code.
- `id:abc123` – Jupyter cell_id field.

Examples:

```bash
nbq section --query match:"## Crossref" --sets 2 < notebook.ipynb
nbq cells --query start:40 --sets 1 --format py < notebook.ipynb
nbq outputs --query id:1234abcd --outputs mime=image/png < notebook.ipynb > screenshot.png
```
