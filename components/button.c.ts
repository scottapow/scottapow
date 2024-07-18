import { cva } from "../styled-system/css";

const styles = cva({
  base: {
    display: 'flex',
  },
  variants: {
    size: {
      lg: {
        paddingBlock: 3,
        paddingInline: 4,
        fontSize: '1.5rem',
      },
      md: {
        paddingBlock: 2,
        paddingInline: 3,
        fontSize: '1.25rem',
      },
      sm: {
        paddingBlock: 1,
        paddingInline: 2,
        fontSize: '1rem',
      }
    },
    color: {
      primary: {
        backgroundColor: 'red.300',
        color: 'white'
      }
    }
  },
  defaultVariants: {
    size: 'sm',
    color: 'primary',
  }
})

type SizeOptions = (typeof styles)['variantMap']['size'][0]
type ColorOptions = (typeof styles)['variantMap']['color'][0]

class CButton extends HTMLButtonElement {
  constructor() {
    super();

    const template = document.createElement('template');
    const size = this.getAttribute('size') as SizeOptions;
    const color = this.getAttribute('color') as ColorOptions;
    template.innerHTML = /* html */`
      <button class=${styles({ size, color })}>
        im a button
      </button>
    `
    let templateContent = template.content;

    const shadowRoot = this.attachShadow({ mode: "open" });
    shadowRoot.appendChild(templateContent.cloneNode(true));
  }

  static observedAttributes = ["size", "color"];

  connectedCallback() {
    console.log("Custom element added to page.");
  }

  disconnectedCallback() {
    console.log("Custom element removed from page.");
  }

  adoptedCallback() {
    console.log("Custom element moved to new page.");
  }

  attributeChangedCallback(name: string, oldValue: string, newValue: string) {
    console.log(`Attribute ${name} has changed.`, { oldValue, newValue });
  }
}

export default function registerCButton() {
  customElements.define("c-button", CButton, { extends: "button" });
};