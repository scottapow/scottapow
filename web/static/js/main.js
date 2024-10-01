"use strict";
(() => {
  // js/styled-system/helpers.js
  function isObject(value) {
    return typeof value === "object" && value != null && !Array.isArray(value);
  }
  function compact(value) {
    return Object.fromEntries(Object.entries(value ?? {}).filter(([_, value2]) => value2 !== void 0));
  }
  var isBaseCondition = (v) => v === "base";
  function filterBaseConditions(c) {
    return c.slice().filter((v) => !isBaseCondition(v));
  }
  function toChar(code) {
    return String.fromCharCode(code + (code > 25 ? 39 : 97));
  }
  function toName(code) {
    let name = "";
    let x;
    for (x = Math.abs(code); x > 52; x = x / 52 | 0)
      name = toChar(x % 52) + name;
    return toChar(x % 52) + name;
  }
  function toPhash(h, x) {
    let i = x.length;
    while (i)
      h = h * 33 ^ x.charCodeAt(--i);
    return h;
  }
  function toHash(value) {
    return toName(toPhash(5381, value) >>> 0);
  }
  var importantRegex = /\s*!(important)?/i;
  function isImportant(value) {
    return typeof value === "string" ? importantRegex.test(value) : false;
  }
  function withoutImportant(value) {
    return typeof value === "string" ? value.replace(importantRegex, "").trim() : value;
  }
  function withoutSpace(str) {
    return typeof str === "string" ? str.replaceAll(" ", "_") : str;
  }
  var memo = (fn) => {
    const cache = /* @__PURE__ */ new Map();
    const get = (...args) => {
      const key = JSON.stringify(args);
      if (cache.has(key)) {
        return cache.get(key);
      }
      const result = fn(...args);
      cache.set(key, result);
      return result;
    };
    return get;
  };
  function mergeProps(...sources) {
    const objects = sources.filter(Boolean);
    return objects.reduce((prev, obj) => {
      Object.keys(obj).forEach((key) => {
        const prevValue = prev[key];
        const value = obj[key];
        if (isObject(prevValue) && isObject(value)) {
          prev[key] = mergeProps(prevValue, value);
        } else {
          prev[key] = value;
        }
      });
      return prev;
    }, {});
  }
  var isNotNullish = (element) => element != null;
  function walkObject(target, predicate, options = {}) {
    const { stop, getKey } = options;
    function inner(value, path = []) {
      if (isObject(value) || Array.isArray(value)) {
        const result = {};
        for (const [prop, child] of Object.entries(value)) {
          const key = getKey?.(prop, child) ?? prop;
          const childPath = [...path, key];
          if (stop?.(value, childPath)) {
            return predicate(value, path);
          }
          const next = inner(child, childPath);
          if (isNotNullish(next)) {
            result[key] = next;
          }
        }
        return result;
      }
      return predicate(value, path);
    }
    return inner(target);
  }
  function toResponsiveObject(values, breakpoints) {
    return values.reduce(
      (acc, current, index) => {
        const key = breakpoints[index];
        if (current != null) {
          acc[key] = current;
        }
        return acc;
      },
      {}
    );
  }
  function normalizeStyleObject(styles2, context2, shorthand = true) {
    const { utility, conditions: conditions2 } = context2;
    const { hasShorthand, resolveShorthand: resolveShorthand2 } = utility;
    return walkObject(
      styles2,
      (value) => {
        return Array.isArray(value) ? toResponsiveObject(value, conditions2.breakpoints.keys) : value;
      },
      {
        stop: (value) => Array.isArray(value),
        getKey: shorthand ? (prop) => hasShorthand ? resolveShorthand2(prop) : prop : void 0
      }
    );
  }
  var fallbackCondition = {
    shift: (v) => v,
    finalize: (v) => v,
    breakpoints: { keys: [] }
  };
  var sanitize = (value) => typeof value === "string" ? value.replaceAll(/[\n\s]+/g, " ") : value;
  function createCss(context2) {
    const { utility, hash, conditions: conds = fallbackCondition } = context2;
    const formatClassName = (str) => [utility.prefix, str].filter(Boolean).join("-");
    const hashFn = (conditions2, className) => {
      let result;
      if (hash) {
        const baseArray = [...conds.finalize(conditions2), className];
        result = formatClassName(utility.toHash(baseArray, toHash));
      } else {
        const baseArray = [...conds.finalize(conditions2), formatClassName(className)];
        result = baseArray.join(":");
      }
      return result;
    };
    return memo(({ base, ...styles2 } = {}) => {
      const styleObject = Object.assign(styles2, base);
      const normalizedObject = normalizeStyleObject(styleObject, context2);
      const classNames = /* @__PURE__ */ new Set();
      walkObject(normalizedObject, (value, paths) => {
        const important = isImportant(value);
        if (value == null)
          return;
        const [prop, ...allConditions] = conds.shift(paths);
        const conditions2 = filterBaseConditions(allConditions);
        const transformed = utility.transform(prop, withoutImportant(sanitize(value)));
        let className = hashFn(conditions2, transformed.className);
        if (important)
          className = `${className}!`;
        classNames.add(className);
      });
      return Array.from(classNames).join(" ");
    });
  }
  function compactStyles(...styles2) {
    return styles2.flat().filter((style) => isObject(style) && Object.keys(compact(style)).length > 0);
  }
  function createMergeCss(context2) {
    function resolve(styles2) {
      const allStyles = compactStyles(...styles2);
      if (allStyles.length === 1)
        return allStyles;
      return allStyles.map((style) => normalizeStyleObject(style, context2));
    }
    function mergeCss2(...styles2) {
      return mergeProps(...resolve(styles2));
    }
    function assignCss2(...styles2) {
      return Object.assign({}, ...resolve(styles2));
    }
    return { mergeCss: memo(mergeCss2), assignCss: assignCss2 };
  }
  var wordRegex = /([A-Z])/g;
  var msRegex = /^ms-/;
  var hypenateProperty = memo((property) => {
    if (property.startsWith("--"))
      return property;
    return property.replace(wordRegex, "-$1").replace(msRegex, "-ms-").toLowerCase();
  });
  var fns = ["min", "max", "clamp", "calc"];
  var fnRegExp = new RegExp(`^(${fns.join("|")})\\(.*\\)`);
  var lengthUnits = "cm,mm,Q,in,pc,pt,px,em,ex,ch,rem,lh,rlh,vw,vh,vmin,vmax,vb,vi,svw,svh,lvw,lvh,dvw,dvh,cqw,cqh,cqi,cqb,cqmin,cqmax,%";
  var lengthUnitsPattern = `(?:${lengthUnits.split(",").join("|")})`;
  var lengthRegExp = new RegExp(`^[+-]?[0-9]*.?[0-9]+(?:[eE][+-]?[0-9]+)?${lengthUnitsPattern}$`);
  function splitProps(props, ...keys) {
    const descriptors = Object.getOwnPropertyDescriptors(props);
    const dKeys = Object.keys(descriptors);
    const split = (k) => {
      const clone = {};
      for (let i = 0; i < k.length; i++) {
        const key = k[i];
        if (descriptors[key]) {
          Object.defineProperty(clone, key, descriptors[key]);
          delete descriptors[key];
        }
      }
      return clone;
    };
    const fn = (key) => split(Array.isArray(key) ? key : dKeys.filter(key));
    return keys.map(fn).concat(split(dKeys));
  }
  var uniq = (...items) => items.filter(Boolean).reduce((acc, item) => Array.from(/* @__PURE__ */ new Set([...acc, ...item])), []);

  // js/styled-system/css/conditions.js
  var conditionsStr = "_hover,_focus,_focusWithin,_focusVisible,_disabled,_active,_visited,_target,_readOnly,_readWrite,_empty,_checked,_enabled,_expanded,_highlighted,_before,_after,_firstLetter,_firstLine,_marker,_selection,_file,_backdrop,_first,_last,_only,_even,_odd,_firstOfType,_lastOfType,_onlyOfType,_peerFocus,_peerHover,_peerActive,_peerFocusWithin,_peerFocusVisible,_peerDisabled,_peerChecked,_peerInvalid,_peerExpanded,_peerPlaceholderShown,_groupFocus,_groupHover,_groupActive,_groupFocusWithin,_groupFocusVisible,_groupDisabled,_groupChecked,_groupExpanded,_groupInvalid,_indeterminate,_required,_valid,_invalid,_autofill,_inRange,_outOfRange,_placeholder,_placeholderShown,_pressed,_selected,_default,_optional,_open,_closed,_fullscreen,_loading,_currentPage,_currentStep,_motionReduce,_motionSafe,_print,_landscape,_portrait,_dark,_light,_osDark,_osLight,_highContrast,_lessContrast,_moreContrast,_ltr,_rtl,_scrollbar,_scrollbarThumb,_scrollbarTrack,_horizontal,_vertical,_starting,sm,smOnly,smDown,md,mdOnly,mdDown,lg,lgOnly,lgDown,xl,xlOnly,xlDown,2xl,2xlOnly,2xlDown,smToMd,smToLg,smToXl,smTo2xl,mdToLg,mdToXl,mdTo2xl,lgToXl,lgTo2xl,xlTo2xl,@/xs,@/sm,@/md,@/lg,@/xl,@/2xl,@/3xl,@/4xl,@/5xl,@/6xl,@/7xl,@/8xl,base";
  var conditions = new Set(conditionsStr.split(","));
  function isCondition(value) {
    return conditions.has(value) || /^@|&|&$/.test(value);
  }
  var underscoreRegex = /^_/;
  var conditionsSelectorRegex = /&|@/;
  function finalizeConditions(paths) {
    return paths.map((path) => {
      if (conditions.has(path)) {
        return path.replace(underscoreRegex, "");
      }
      if (conditionsSelectorRegex.test(path)) {
        return `[${withoutSpace(path.trim())}]`;
      }
      return path;
    });
  }
  function sortConditions(paths) {
    return paths.sort((a, b) => {
      const aa = isCondition(a);
      const bb = isCondition(b);
      if (aa && !bb) return 1;
      if (!aa && bb) return -1;
      return 0;
    });
  }

  // js/styled-system/css/css.js
  var utilities = "aspectRatio:aspect,boxDecorationBreak:decoration,zIndex:z,boxSizing:box,objectPosition:obj-pos,objectFit:obj-fit,overscrollBehavior:overscroll,overscrollBehaviorX:overscroll-x,overscrollBehaviorY:overscroll-y,position:pos/1,top:top,left:left,insetInline:inset-x/insetX,insetBlock:inset-y/insetY,inset:inset,insetBlockEnd:inset-b,insetBlockStart:inset-t,insetInlineEnd:end/insetEnd/1,insetInlineStart:start/insetStart/1,right:right,bottom:bottom,float:float,visibility:vis,display:d,hideFrom:hide,hideBelow:show,flexBasis:basis,flex:flex,flexDirection:flex/flexDir,flexGrow:grow,flexShrink:shrink,gridTemplateColumns:grid-cols,gridTemplateRows:grid-rows,gridColumn:col-span,gridRow:row-span,gridColumnStart:col-start,gridColumnEnd:col-end,gridAutoFlow:grid-flow,gridAutoColumns:auto-cols,gridAutoRows:auto-rows,gap:gap,gridGap:gap,gridRowGap:gap-x,gridColumnGap:gap-y,rowGap:gap-x,columnGap:gap-y,justifyContent:justify,alignContent:content,alignItems:items,alignSelf:self,padding:p/1,paddingLeft:pl/1,paddingRight:pr/1,paddingTop:pt/1,paddingBottom:pb/1,paddingBlock:py/1/paddingY,paddingBlockEnd:pb,paddingBlockStart:pt,paddingInline:px/paddingX/1,paddingInlineEnd:pe/1/paddingEnd,paddingInlineStart:ps/1/paddingStart,marginLeft:ml/1,marginRight:mr/1,marginTop:mt/1,marginBottom:mb/1,margin:m/1,marginBlock:my/1/marginY,marginBlockEnd:mb,marginBlockStart:mt,marginInline:mx/1/marginX,marginInlineEnd:me/1/marginEnd,marginInlineStart:ms/1/marginStart,spaceX:space-x,spaceY:space-y,outlineWidth:ring-width/ringWidth,outlineColor:ring-color/ringColor,outline:ring/1,outlineOffset:ring-offset/ringOffset,divideX:divide-x,divideY:divide-y,divideColor:divide-color,divideStyle:divide-style,width:w/1,inlineSize:w,minWidth:min-w/minW,minInlineSize:min-w,maxWidth:max-w/maxW,maxInlineSize:max-w,height:h/1,blockSize:h,minHeight:min-h/minH,minBlockSize:min-h,maxHeight:max-h/maxH,maxBlockSize:max-b,color:text,fontFamily:font,fontSize:fs,fontWeight:fw,fontSmoothing:smoothing,fontVariantNumeric:numeric,letterSpacing:tracking,lineHeight:leading,textAlign:text-align,textDecoration:text-decor,textDecorationColor:text-decor-color,textEmphasisColor:text-emphasis-color,textDecorationStyle:decoration-style,textDecorationThickness:decoration-thickness,textUnderlineOffset:underline-offset,textTransform:text-transform,textIndent:indent,textShadow:text-shadow,textShadowColor:text-shadow/textShadowColor,textOverflow:text-overflow,verticalAlign:v-align,wordBreak:break,textWrap:text-wrap,truncate:truncate,lineClamp:clamp,listStyleType:list-type,listStylePosition:list-pos,listStyleImage:list-img,backgroundPosition:bg-pos/bgPosition,backgroundPositionX:bg-pos-x/bgPositionX,backgroundPositionY:bg-pos-y/bgPositionY,backgroundAttachment:bg-attach/bgAttachment,backgroundClip:bg-clip/bgClip,background:bg/1,backgroundColor:bg/bgColor,backgroundOrigin:bg-origin/bgOrigin,backgroundImage:bg-img/bgImage,backgroundRepeat:bg-repeat/bgRepeat,backgroundBlendMode:bg-blend/bgBlendMode,backgroundSize:bg-size/bgSize,backgroundGradient:bg-gradient/bgGradient,textGradient:text-gradient,gradientFromPosition:gradient-from-pos,gradientToPosition:gradient-to-pos,gradientFrom:gradient-from,gradientTo:gradient-to,gradientVia:gradient-via,gradientViaPosition:gradient-via-pos,borderRadius:rounded/1,borderTopLeftRadius:rounded-tl/roundedTopLeft,borderTopRightRadius:rounded-tr/roundedTopRight,borderBottomRightRadius:rounded-br/roundedBottomRight,borderBottomLeftRadius:rounded-bl/roundedBottomLeft,borderTopRadius:rounded-t/roundedTop,borderRightRadius:rounded-r/roundedRight,borderBottomRadius:rounded-b/roundedBottom,borderLeftRadius:rounded-l/roundedLeft,borderStartStartRadius:rounded-ss/roundedStartStart,borderStartEndRadius:rounded-se/roundedStartEnd,borderStartRadius:rounded-s/roundedStart,borderEndStartRadius:rounded-es/roundedEndStart,borderEndEndRadius:rounded-ee/roundedEndEnd,borderEndRadius:rounded-e/roundedEnd,border:border,borderWidth:border-w,borderTopWidth:border-tw,borderLeftWidth:border-lw,borderRightWidth:border-rw,borderBottomWidth:border-bw,borderColor:border,borderInline:border-x/borderX,borderInlineWidth:border-x/borderXWidth,borderInlineColor:border-x/borderXColor,borderBlock:border-y/borderY,borderBlockWidth:border-y/borderYWidth,borderBlockColor:border-y/borderYColor,borderLeft:border-l,borderLeftColor:border-l,borderInlineStart:border-s/borderStart,borderInlineStartWidth:border-s/borderStartWidth,borderInlineStartColor:border-s/borderStartColor,borderRight:border-r,borderRightColor:border-r,borderInlineEnd:border-e/borderEnd,borderInlineEndWidth:border-e/borderEndWidth,borderInlineEndColor:border-e/borderEndColor,borderTop:border-t,borderTopColor:border-t,borderBottom:border-b,borderBottomColor:border-b,borderBlockEnd:border-be,borderBlockEndColor:border-be,borderBlockStart:border-bs,borderBlockStartColor:border-bs,boxShadow:shadow/1,boxShadowColor:shadow-color/shadowColor,mixBlendMode:mix-blend,filter:filter,brightness:brightness,contrast:contrast,grayscale:grayscale,hueRotate:hue-rotate,invert:invert,saturate:saturate,sepia:sepia,dropShadow:drop-shadow,blur:blur,backdropFilter:backdrop,backdropBlur:backdrop-blur,backdropBrightness:backdrop-brightness,backdropContrast:backdrop-contrast,backdropGrayscale:backdrop-grayscale,backdropHueRotate:backdrop-hue-rotate,backdropInvert:backdrop-invert,backdropOpacity:backdrop-opacity,backdropSaturate:backdrop-saturate,backdropSepia:backdrop-sepia,borderCollapse:border,borderSpacing:border-spacing,borderSpacingX:border-spacing-x,borderSpacingY:border-spacing-y,tableLayout:table,transitionTimingFunction:ease,transitionDelay:delay,transitionDuration:duration,transitionProperty:transition-prop,transition:transition,animation:animation,animationName:animation-name,animationTimingFunction:animation-ease,animationDuration:animation-duration,animationDelay:animation-delay,transformOrigin:origin,rotate:rotate,rotateX:rotate-x,rotateY:rotate-y,rotateZ:rotate-z,scale:scale,scaleX:scale-x,scaleY:scale-y,translate:translate,translateX:translate-x/x,translateY:translate-y/y,translateZ:translate-z/z,accentColor:accent,caretColor:caret,scrollBehavior:scroll,scrollbar:scrollbar,scrollMargin:scroll-m,scrollMarginLeft:scroll-ml,scrollMarginRight:scroll-mr,scrollMarginTop:scroll-mt,scrollMarginBottom:scroll-mb,scrollMarginBlock:scroll-my/scrollMarginY,scrollMarginBlockEnd:scroll-mb,scrollMarginBlockStart:scroll-mt,scrollMarginInline:scroll-mx/scrollMarginX,scrollMarginInlineEnd:scroll-me,scrollMarginInlineStart:scroll-ms,scrollPadding:scroll-p,scrollPaddingBlock:scroll-pb/scrollPaddingY,scrollPaddingBlockStart:scroll-pt,scrollPaddingBlockEnd:scroll-pb,scrollPaddingInline:scroll-px/scrollPaddingX,scrollPaddingInlineEnd:scroll-pe,scrollPaddingInlineStart:scroll-ps,scrollPaddingLeft:scroll-pl,scrollPaddingRight:scroll-pr,scrollPaddingTop:scroll-pt,scrollPaddingBottom:scroll-pb,scrollSnapAlign:snap-align,scrollSnapStop:snap-stop,scrollSnapType:snap-type,scrollSnapStrictness:snap-strictness,scrollSnapMargin:snap-m,scrollSnapMarginTop:snap-mt,scrollSnapMarginBottom:snap-mb,scrollSnapMarginLeft:snap-ml,scrollSnapMarginRight:snap-mr,touchAction:touch,userSelect:select,fill:fill,stroke:stroke,strokeWidth:stroke-w,srOnly:sr,debug:debug,appearance:appearance,backfaceVisibility:backface,clipPath:clip-path,hyphens:hyphens,mask:mask,maskImage:mask-image,maskSize:mask-size,textSizeAdjust:text-adjust,container:cq,containerName:cq-name,containerType:cq-type,textStyle:textStyle";
  var classNameByProp = /* @__PURE__ */ new Map();
  var shorthands = /* @__PURE__ */ new Map();
  utilities.split(",").forEach((utility) => {
    const [prop, meta] = utility.split(":");
    const [className, ...shorthandList] = meta.split("/");
    classNameByProp.set(prop, className);
    if (shorthandList.length) {
      shorthandList.forEach((shorthand) => {
        shorthands.set(shorthand === "1" ? className : shorthand, prop);
      });
    }
  });
  var resolveShorthand = (prop) => shorthands.get(prop) || prop;
  var context = {
    conditions: {
      shift: sortConditions,
      finalize: finalizeConditions,
      breakpoints: { keys: ["base", "sm", "md", "lg", "xl", "2xl"] }
    },
    utility: {
      transform: (prop, value) => {
        const key = resolveShorthand(prop);
        const propKey = classNameByProp.get(key) || hypenateProperty(key);
        return { className: `${propKey}_${withoutSpace(value)}` };
      },
      hasShorthand: true,
      toHash: (path, hashFn) => hashFn(path.join(":")),
      resolveShorthand
    }
  };
  var cssFn = createCss(context);
  var css = (...styles2) => cssFn(mergeCss(...styles2));
  css.raw = (...styles2) => mergeCss(...styles2);
  var { mergeCss, assignCss } = createMergeCss(context);

  // js/styled-system/css/cva.js
  var defaults = (conf) => ({
    base: {},
    variants: {},
    defaultVariants: {},
    compoundVariants: [],
    ...conf
  });
  function cva(config) {
    const { base, variants, defaultVariants, compoundVariants } = defaults(config);
    const getVariantProps = (variants2) => ({ ...defaultVariants, ...compact(variants2) });
    function resolve(props = {}) {
      const computedVariants = getVariantProps(props);
      let variantCss = { ...base };
      for (const [key, value] of Object.entries(computedVariants)) {
        if (variants[key]?.[value]) {
          variantCss = mergeCss(variantCss, variants[key][value]);
        }
      }
      const compoundVariantCss = getCompoundVariantCss(compoundVariants, computedVariants);
      return mergeCss(variantCss, compoundVariantCss);
    }
    function merge(__cva) {
      const override = defaults(__cva.config);
      const variantKeys2 = uniq(__cva.variantKeys, Object.keys(variants));
      return cva({
        base: mergeCss(base, override.base),
        variants: Object.fromEntries(
          variantKeys2.map((key) => [key, mergeCss(variants[key], override.variants[key])])
        ),
        defaultVariants: mergeProps(defaultVariants, override.defaultVariants),
        compoundVariants: [...compoundVariants, ...override.compoundVariants]
      });
    }
    function cvaFn(props) {
      return css(resolve(props));
    }
    const variantKeys = Object.keys(variants);
    function splitVariantProps(props) {
      return splitProps(props, variantKeys);
    }
    const variantMap = Object.fromEntries(Object.entries(variants).map(([key, value]) => [key, Object.keys(value)]));
    return Object.assign(memo(cvaFn), {
      __cva__: true,
      variantMap,
      variantKeys,
      raw: resolve,
      config,
      merge,
      splitVariantProps,
      getVariantProps
    });
  }
  function getCompoundVariantCss(compoundVariants, variantMap) {
    let result = {};
    compoundVariants.forEach((compoundVariant) => {
      const isMatching = Object.entries(compoundVariant).every(([key, value]) => {
        if (key === "css") return true;
        const values = Array.isArray(value) ? value : [value];
        return values.some((value2) => variantMap[key] === value2);
      });
      if (isMatching) {
        result = mergeCss(result, compoundVariant.css);
      }
    });
    return result;
  }

  // js/button.c.ts
  var styles = cva({
    base: {
      display: "flex"
    },
    variants: {
      size: {
        lg: {
          paddingBlock: 3,
          paddingInline: 4,
          fontSize: "1.5rem"
        },
        md: {
          paddingBlock: 2,
          paddingInline: 3,
          fontSize: "1.25rem"
        },
        sm: {
          paddingBlock: 1,
          paddingInline: 2,
          fontSize: "1rem"
        }
      },
      color: {
        primary: {
          backgroundColor: "red.300",
          color: "white"
        }
      }
    },
    defaultVariants: {
      size: "sm",
      color: "primary"
    }
  });
  var CButton = class extends HTMLButtonElement {
    constructor() {
      super();
      const size = this.getAttribute("size") ?? "md";
      const color = this.getAttribute("color") ?? "primary";
      this.classList.add(...styles({ size, color }).split(" "));
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
    attributeChangedCallback(name, oldValue, newValue) {
      console.log(`Attribute ${name} has changed.`, { oldValue, newValue });
    }
  };
  function registerCButton() {
    customElements.define("wc-button", CButton, { extends: "button" });
  }

  // js/main.ts
  registerCButton();
})();
