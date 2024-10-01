import { cva } from "./styled-system/css/index.js";

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

    const size = (this.getAttribute('size') ?? 'md') as SizeOptions;
    const color = (this.getAttribute('color') ?? 'primary') as ColorOptions;
    this.classList.add(...styles({ size, color }).split(' '))
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
  customElements.define("wc-button", CButton, { extends: "button" });
};