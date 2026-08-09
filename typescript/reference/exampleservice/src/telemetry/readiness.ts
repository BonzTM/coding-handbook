export class Readiness {
  #ready = false;

  markReady(): void {
    this.#ready = true;
  }

  markDraining(): void {
    this.#ready = false;
  }

  isReady(): boolean {
    return this.#ready;
  }
}
