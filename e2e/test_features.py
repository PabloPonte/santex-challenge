import json
import uuid
from urllib import error, request


def http_json(method, url, payload=None):
    body = None if payload is None else json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"} if body else {}
    http_request = request.Request(url, data=body, headers=headers, method=method)
    try:
        with request.urlopen(http_request, timeout=5) as response:
            return response.status, json.loads(response.read() or b"{}")
    except error.HTTPError as exc:
        return exc.code, json.loads(exc.read() or b"{}")


def test_crud_and_cache_only_evaluation(base_url):
    name = f"e2e-{uuid.uuid4().hex}"
    feature_url = f"{base_url}/api/v1/internal/features/{name}"
    external_url = f"{base_url}/api/v1/external/features/{name}/users/allowed-user/evaluation"
    payload = {
        "name": name,
        "description": "E2E feature",
        "status": "whitelisted",
        "whitelist": ["allowed-user"],
    }

    try:
        status, created = http_json("POST", f"{base_url}/api/v1/internal/features", payload)
        assert status == 201
        assert created["name"] == name
        assert created["statusDate"]

        status, evaluation = http_json("GET", external_url)
        assert status == 200
        assert evaluation == {
            "featureName": name,
            "userId": "allowed-user",
            "enabled": True,
            "status": "whitelisted",
        }

        update = {"description": "Enabled for everyone", "status": "open", "whitelist": ["allowed-user"]}
        status, updated = http_json("PUT", feature_url, update)
        assert status == 200
        assert updated["status"] == "open"

        status, evaluation = http_json(
            "GET", f"{base_url}/api/v1/external/features/{name}/users/another-user/evaluation"
        )
        assert status == 200
        assert evaluation["enabled"] is True

        status, refreshed = http_json("POST", f"{base_url}/api/v1/internal/cache/refresh")
        assert status == 200
        assert refreshed["refreshed"] >= 1
    finally:
        http_json("DELETE", feature_url)


def test_rejects_invalid_feature_name(base_url):
    status, response = http_json(
        "POST",
        f"{base_url}/api/v1/internal/features",
        {"name": "Invalid Name", "description": "", "status": "open", "whitelist": []},
    )

    assert status == 400
    assert "validation" in response["error"]
