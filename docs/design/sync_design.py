"""Converte le schermate dell'artefatto di design in pagine HTML statiche per il repo.

L'artefatto "SpesTube" su claude.ai è la fonte del design; questa cartella ne tiene una copia
leggibile senza claude.ai e senza server. Lo script non scarica niente: riceve la cartella
in cui sono stati salvati `canvas.json` e i file `*.dc.html` dell'artefatto (lo fa il comando
`/sync-design` di Claude Code) e scrive qui una pagina per schermata, nell'ordine del canvas.

Uso:
    python3 -I docs/design/sync_design.py <cartella-scaricata> docs/design
"""

import html
import json
import pathlib
import re
import sys

# La prima schermata del canvas è la presentazione del progetto: qui ha un nome che lo dice.
RENAMES = {"Main": "progetto"}

README_START = "<!-- schermate:inizio -->"
README_END = "<!-- schermate:fine -->"


def page_name(dc_file: str) -> str:
    stem = dc_file.removesuffix(".dc.html")
    if stem in RENAMES:
        return RENAMES[stem] + ".html"
    # OggiScuro -> oggi-scuro, DaVedere -> da-vedere
    return re.sub(r"(?<!^)(?=[A-Z])", "-", stem).lower() + ".html"


def convert(source: str, names: dict[str, str], phone: bool) -> str:
    title = re.search(r"<title>(.*?)</title>", source, re.S)
    helmet = re.search(r"<helmet>(.*?)</helmet>", source, re.S)
    body = re.search(r"</helmet>(.*?)</x-dc>", source, re.S)
    if not (title and helmet and body):
        raise ValueError("forma della schermata non riconosciuta")
    head, content = helmet.group(1).strip(), body.group(1).strip()
    # I collegamenti fra schermate puntano alle pagine statiche, non ai file del canvas.
    for dc_file, name in names.items():
        content = content.replace(f'href="{dc_file}"', f'href="{name}"')
    if phone:
        # Una schermata da telefono si mostra al centro, su un fondo neutro.
        head = head.replace(
            "body{margin:0}",
            "body{margin:0}\nhtml{background:#D9DCE3}\nbody{display:flex;justify-content:center;padding:24px 0}",
        )
    return f"""<!doctype html>
<html lang="it">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<!-- Generata da sync_design.py a partire dall'artefatto di design: non modificare a mano. -->
<title>{title.group(1)}</title>
{head}
</head>
<body>
{content}
</body>
</html>
"""


def main(source_dir: pathlib.Path, target_dir: pathlib.Path) -> None:
    canvas = json.loads((source_dir / "canvas.json").read_text())
    order = canvas["order"]
    boards = canvas["boards"]
    names = {dc_file: page_name(dc_file) for dc_file in order}

    written = []
    for dc_file in order:
        source = (source_dir / dc_file).read_text()
        phone = boards[dc_file].get("expand") != "fill"
        (target_dir / names[dc_file]).write_text(convert(source, names, phone))
        written.append(dc_file)

    # Una schermata tolta dal canvas sparisce anche da qui.
    for stale in sorted(set(p.name for p in target_dir.glob("*.html")) - set(names.values())):
        (target_dir / stale).unlink()
        print("tolta", stale)

    rows = ["| File | Schermata |", "| --- | --- |"]
    for dc_file in order:
        label = boards[dc_file].get("title") or dc_file.removesuffix(".dc.html")
        rows.append(f"| [`{names[dc_file]}`](./{names[dc_file]}) | {html.unescape(label)} |")
    readme = target_dir / "README.md"
    text = readme.read_text()
    start, end = text.index(README_START), text.index(README_END)
    readme.write_text(text[: start + len(README_START)] + "\n\n" + "\n".join(rows) + "\n\n" + text[end:])

    for dc_file in written:
        print("scritta", names[dc_file])


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    main(pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]))
