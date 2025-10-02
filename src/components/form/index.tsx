import FormField from "./form-field.astro";
import FormLabel from "./form-label.astro";
import FormInput from "./form-input.astro";
import FormSelect from "./form-select.astro";
import FormActions from "./form-actions.astro";
import _Form from "./form.astro";

type FormModule = typeof _Form & {
  Field: typeof FormField;
  Label: typeof FormLabel;
  Input: typeof FormInput;
  Select: typeof FormSelect;
  Actions: typeof FormActions;
};

const Form = _Form as FormModule;
Form.Field = FormField;
Form.Label = FormLabel;
Form.Input = FormInput;
Form.Select = FormSelect;
Form.Actions = FormActions;

export default Form;