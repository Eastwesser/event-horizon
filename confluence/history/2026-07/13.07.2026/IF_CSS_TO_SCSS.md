В ваших компонентах должны быть импорты .scss:

tsx
// src/components/Shop/Shop.tsx
import './Shop.css'; // или './scss/Shop.scss' если используете SCSS
Если вы хотите использовать SCSS, измените импорты на:

tsx
import './scss/Shop.scss';
import './scss/ShopItemCard.scss';
import './scss/PurchaseModal.scss';