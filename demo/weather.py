"""Weather agent with live data from wttr.in. uv run uvicorn weather:app --port 9201"""

import os
from datetime import date

import httpx

from util import agent_app

NAME = "weather"
VERSION = "1.0.0"
PORT = int(os.environ.get("PORT", "9201"))
CONDITIONS = ["sunny", "clear", "cloudy", "rain", "snow", "windy", "fog"]

RAIN = {176, 200, 263, 266, 293, 296, 299, 302, 305, 308, 311, 353, 356, 359, 362, 365, 386, 389, 392, 395}
SNOW = {179, 182, 227, 230, 317, 320, 323, 326, 329, 332, 335, 338, 350, 371, 374, 377}
FOG = {143, 248, 260}
CLOUDY = {116, 119, 122}


def condition(code, wind_kph=0, is_day=True):
    code = int(code)

    if code in SNOW:
        return "snow"

    if code in RAIN:
        return "rain"

    if code in FOG:
        return "fog"

    if code in CLOUDY:
        return "cloudy"

    if wind_kph >= 40:
        return "windy"

    return "sunny" if is_day else "clear"


def report(city):
    response = httpx.get(f"https://wttr.in/{city}?format=j1", timeout=10)
    response.raise_for_status()

    return response.json()


def current(args):
    city = args.get("city") or "Berlin"
    data = report(city)
    now = data["current_condition"][0]
    wind = int(now["windspeedKmph"])
    is_day = "night" not in now["weatherIconUrl"][0]["value"].lower()

    return {
        "city": data["nearest_area"][0]["areaName"][0]["value"],
        "temperature_c": int(now["temp_C"]),
        "humidity": int(now["humidity"]),
        "wind_kph": wind,
        "condition": condition(now["weatherCode"], wind, is_day),
    }


def forecast(args):
    city = args.get("city") or "Berlin"
    days = int(args.get("days") or 3)
    data = report(city)

    return {
        "city": data["nearest_area"][0]["areaName"][0]["value"],
        "days": [
            {
                "day": date.fromisoformat(entry["date"]).strftime("%a").lower(),
                "high_c": int(entry["maxtempC"]),
                "low_c": int(entry["mintempC"]),
                "condition": condition(entry["hourly"][4]["weatherCode"]),
            }
            for entry in data["weather"][:days]
        ],
    }


def alerts(args):
    city = args.get("city") or "Berlin"
    geo = httpx.get(
        "https://geocoding-api.open-meteo.com/v1/search",
        params={"name": city, "count": 1},
        timeout=10,
    ).json()["results"][0]

    # weather.gov covers the US; anywhere else has no alerts, which is an empty list.
    active = httpx.get(
        f"https://api.weather.gov/alerts/active?point={geo['latitude']},{geo['longitude']}",
        headers={"User-Agent": "sloth-demo/0.1", "Accept": "application/geo+json"},
        timeout=10,
    ).json()

    return {
        "city": geo["name"],
        "alerts": [
            {"type": item["properties"]["event"], "severity": item["properties"]["severity"].lower()}
            for item in active["features"]
        ],
    }


app = agent_app(
    NAME,
    "Weather",
    "Current conditions, forecasts and alerts for a city",
    VERSION,
    PORT,
    [
        {
            "id": "weather.current.get",
            "name": "Current weather",
            "description": "Return temperature, humidity, wind and condition for a city. Data: {\"city\": \"Berlin\"}",
            "tags": ["weather", "current"],
            "inputModes": ["application/json"],
            "examples": ['{"city": "Berlin"}'],
            "func": current,
        },
        {
            "id": "weather.forecast.get",
            "name": "Weather forecast",
            "description": "Return a daily high, low and condition forecast for a city. Data: {\"city\": \"Berlin\", \"days\": 3}",
            "tags": ["weather", "forecast"],
            "inputModes": ["application/json"],
            "examples": ['{"city": "Cairo", "days": 5}'],
            "func": forecast,
        },
        {
            "id": "weather.alerts.list",
            "name": "Weather alerts",
            "description": "List active severe weather alerts for a city. Data: {\"city\": \"Miami\"}",
            "tags": ["weather", "alerts"],
            "inputModes": ["application/json"],
            "examples": ['{"city": "Miami"}'],
            "func": alerts,
        },
    ],
)
