export class Readiness {
  #ready = false;

  set(ready: boolean): void {
    this.#ready = ready;
  }

  isReady(): boolean {
    return this.#ready;
  }
}
