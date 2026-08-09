/** @type {import("jest").Config} */
if (process.env.RUN_INTEGRATION !== "1") {
  throw new Error("integration tests require RUN_INTEGRATION=1");
}

module.exports = {
  testEnvironment: "node",
  testMatch: ["<rootDir>/src/integration/**/*.integration.test.ts"],
  transform: { "^.+\\.tsx?$": "babel-jest" },
  moduleNameMapper: { "^(\\.{1,2}/.*)\\.js$": "$1" },
  clearMocks: true,
  restoreMocks: true,
  maxWorkers: 1,
  testTimeout: 60000,
};
