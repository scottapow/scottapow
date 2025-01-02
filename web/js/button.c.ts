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

class CButton extends HTMLElement {
  constructor() {
    super();
  }

  static observedAttributes = ["size", "color"];

  connectedCallback() {
    let btn = document.createElement('button');
    let size = (this.getAttribute('size') ?? 'md') as SizeOptions;
    let color = (this.getAttribute('color') ?? 'primary') as ColorOptions;
    btn.classList.add(...styles({ size, color }).split(' '));
    btn.replaceChildren(...this.children);
    btn.append(this.innerHTML);
    this.replaceChildren(btn);
  }
}

export default function registerCButton() {
  customElements.define("wc-button", CButton);
};