import { expect, it } from "@jest/globals";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderApp } from "../test/render-app.js";

it("navigates from the list to a lazy widget detail route", async () => {
  const user = userEvent.setup();
  renderApp();

  await user.click(await screen.findByRole("link", { name: "Meter" }));

  expect(
    await screen.findByRole("heading", { name: "Meter", level: 1 }),
  ).toBeVisible();
  expect(screen.getByText("Measures throughput")).toBeVisible();
  expect(document.title).toBe("Meter | Widget administration");
});

it("loads a widget detail from a direct deep link", async () => {
  renderApp("/widgets/6ba7b810-9dad-41d1-80b4-00c04fd430c8");

  expect(
    await screen.findByRole("heading", { name: "Gauge", level: 1 }),
  ).toBeVisible();
  expect(screen.getByText("No description")).toBeVisible();
});

it("rejects a malformed deep-link parameter without requesting it", async () => {
  renderApp("/widgets/not-a-uuid");

  expect(await screen.findByRole("alert")).toHaveTextContent(
    "The widget address is invalid",
  );
});
