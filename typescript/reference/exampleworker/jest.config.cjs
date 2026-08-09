/** @type {import("jest").Config} */
module.exports = {
  testEnvironment: "node",
  testMatch: ["<rootDir>/src/**/*.test.ts"],
  testPathIgnorePatterns: ["/integration/", "\\.integration\\.test\\.ts$"],
  transform: { "^.+\\.tsx?$": "babel-jest" },
  moduleNameMapper: { "^(\\.{1,2}/.*)\\.js$": "$1" },
  clearMocks: true,
  restoreMocks: true,
  maxWorkers: 2,
};
