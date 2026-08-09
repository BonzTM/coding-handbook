import {
  useRef,
  useState,
  type ReactNode,
  type RefObject,
  type SyntheticEvent,
} from "react";
import {
  createWidgetFormSchema,
  type CreateWidgetForm,
} from "../api/widget-schemas.js";
import { useCreateWidget } from "../hooks/widget-queries.js";

type FormErrors = Partial<Record<keyof CreateWidgetForm, string>>;

export function WidgetForm(): ReactNode {
  const mutation = useCreateWidget();
  const [errors, setErrors] = useState<FormErrors>({});
  const nameRef = useRef<HTMLInputElement>(null);

  function handleSubmit(
    event: SyntheticEvent<HTMLFormElement, SubmitEvent>,
  ): void {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const parsed = createWidgetFormSchema.safeParse({
      name: form.get("name"),
      description: form.get("description"),
    });
    if (!parsed.success) {
      setErrors(readErrors(parsed.error));
      nameRef.current?.focus();
      return;
    }
    setErrors({});
    const id = crypto.randomUUID();
    mutation.mutate({
      id,
      idempotencyKey: `create-${id}`,
      name: parsed.data.name,
      description:
        parsed.data.description.length === 0 ? null : parsed.data.description,
    });
  }

  return (
    <section aria-labelledby="create-widget-heading">
      <h2 id="create-widget-heading">Create widget</h2>
      {Object.keys(errors).length > 0 ? (
        <p role="alert">Correct the highlighted fields.</p>
      ) : null}
      {mutation.isError ? (
        <p role="alert">The widget could not be created. Try again.</p>
      ) : null}
      <form onSubmit={handleSubmit} noValidate>
        <WidgetFields errors={errors} nameRef={nameRef} />
        <button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? "Creating…" : "Create widget"}
        </button>
      </form>
    </section>
  );
}

function WidgetFields({
  errors,
  nameRef,
}: Readonly<{
  errors: FormErrors;
  nameRef: RefObject<HTMLInputElement | null>;
}>): ReactNode {
  return (
    <>
      <div>
        <label htmlFor="widget-name">Name</label>
        <input
          id="widget-name"
          ref={nameRef}
          name="name"
          aria-invalid={errors.name === undefined ? undefined : true}
          aria-describedby={
            errors.name === undefined ? undefined : "name-error"
          }
          maxLength={100}
        />
        {errors.name === undefined ? null : (
          <span id="name-error">{errors.name}</span>
        )}
      </div>
      <div>
        <label htmlFor="widget-description">Description</label>
        <textarea
          id="widget-description"
          name="description"
          aria-invalid={errors.description === undefined ? undefined : true}
          aria-describedby={
            errors.description === undefined ? undefined : "description-error"
          }
          maxLength={500}
        />
        {errors.description === undefined ? null : (
          <span id="description-error">{errors.description}</span>
        )}
      </div>
    </>
  );
}

function readErrors(error: {
  readonly issues: readonly {
    readonly path: readonly PropertyKey[];
    readonly message: string;
  }[];
}): FormErrors {
  const errors: FormErrors = {};
  for (const issue of error.issues.slice(0, 10)) {
    const field = issue.path[0];
    if (field === "name" || field === "description") {
      errors[field] ??= issue.message;
    }
  }
  return errors;
}
