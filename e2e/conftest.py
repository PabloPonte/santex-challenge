def pytest_addoption(parser):
    parser.addoption("--base-url", action="store", default="http://localhost:8080")


def pytest_generate_tests(metafunc):
    if "base_url" in metafunc.fixturenames:
        metafunc.parametrize("base_url", [metafunc.config.getoption("base_url")])
