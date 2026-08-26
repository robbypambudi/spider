from __future__ import annotations

from pathlib import Path

import httpx
import typer
import yaml
from pkg.config.settings import get_settings

app = typer.Typer(help="SPIDER control CLI.")
security_app = typer.Typer(help="Security pipeline commands.")
worker_app = typer.Typer(help="Worker inventory commands.")
serving_app = typer.Typer(help="Serving inventory commands.")
inference_app = typer.Typer(help="Protected inference commands.")
app.add_typer(security_app, name="security")
app.add_typer(worker_app, name="worker")
app.add_typer(serving_app, name="serving")
app.add_typer(inference_app, name="inference")


def _client() -> tuple[httpx.Client, dict[str, str]]:
    settings = get_settings()
    token = _login(settings)
    headers = {"Authorization": f"Bearer {token}"}
    return httpx.Client(base_url=settings.api_base_url, timeout=30.0), headers


def _login(settings) -> str:  # type: ignore[no-untyped-def]
    with httpx.Client(base_url=settings.api_base_url, timeout=15.0) as client:
        response = client.post(
            "/api/v1/auth/login",
            json={
                "email": settings.bootstrap_admin_email,
                "password": settings.bootstrap_admin_password,
            },
        )
        response.raise_for_status()
        return str(response.json()["access_token"])


@security_app.command("scan")
def security_scan(text: str) -> None:
    client, headers = _client()
    with client:
        response = client.post("/api/v1/security/scan", json={"text": text}, headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@security_app.command("detectors")
def security_detectors() -> None:
    client, headers = _client()
    with client:
        response = client.get("/api/v1/security/detectors", headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@security_app.command("policies")
def security_policies() -> None:
    client, headers = _client()
    with client:
        response = client.get("/api/v1/security/policies", headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@worker_app.command("list")
def worker_list() -> None:
    client, headers = _client()
    with client:
        response = client.get("/api/v1/workers", headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@worker_app.command("inspect")
def worker_inspect(worker_id: str) -> None:
    client, headers = _client()
    with client:
        response = client.get(f"/api/v1/workers/{worker_id}", headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@serving_app.command("models")
def serving_models() -> None:
    client, headers = _client()
    with client:
        response = client.get("/api/v1/serving/models", headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


@inference_app.command("submit")
def inference_submit(request_file: Path) -> None:
    payload = yaml.safe_load(request_file.read_text(encoding="utf-8"))
    client, headers = _client()
    with client:
        response = client.post("/api/v1/inference", json=payload, headers=headers)
        response.raise_for_status()
        typer.echo(response.text)


def main() -> None:
    app()


if __name__ == "__main__":
    main()
