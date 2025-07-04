<script setup>
import { ref, onMounted } from 'vue';
import { SelectFile, UploadFile, SaveConfig, LoadConfig, Preview } from '../../wailsjs/go/main/App.js';

const type = ref('github');

const githubConf = ref({
  repo: "",
  owner: "",
  token: "", // github write token
  branch: "master",
  prefix: "SECOND",
  commit: "upload by upload_obj",
  author: {
    name: "upload_obj",
    mail: "upload_obj@test.com",
  },
});

const watermarkConf = ref({
  text: '',
  size: 40.0,
  dpi: 100.0,
  color: '#000000',
  x: 100,
  y: 100,
});

const selectedFile = ref('');
const previewFile = ref('');
const previewBlock = ref(false);
const uploadedUrl = ref('');

let previewTimeout = null;

const handleFileSelect = async () => {
  try {
    const filePath = await SelectFile();
    if (filePath) {
      selectedFile.value = filePath;
      preview();
    }
  } catch (error) {
    console.error('Error selecting file:', error);
  }
};

const preview = () => {
  if (!previewTimeout) {
    previewTimeout = setTimeout(previewInner, 1);
    return;
  }

  saveConfig();
  clearTimeout(previewTimeout);
  previewTimeout = setTimeout(previewInner, 1000);
};


const previewInner = async () => {
  try {
    previewBlock.value = true;
    previewFile.value = "";
    previewFile.value = await Preview(selectedFile.value);
    previewBlock.value = false;
  } catch (error) {
    console.error('Error preview file:', error);
    previewBlock.value = false;
  }
}

const uploadFile = async () => {
  try {
    const url = await UploadFile(selectedFile.value, config.value.type);
    uploadedUrl.value = url;
  } catch (error) {
    console.error('Error upload file:', error);
  }
};

const saveConfig = async () => {
  try {
    await SaveConfig(JSON.stringify({
      type: type.value,
      github: githubConf.value,
      watermark: watermarkConf.value,
    }));
  } catch (error) {
    console.error('Error save config:', error);
  }
};


const loadConfig = async () => {
  try {
    const configJson = await LoadConfig();
    const config = JSON.parse(configJson);

    if (config?.type) {
      type.value = config.type;
    }

    if (config?.github) {
      githubConf.value = config.github;
    }

    if (config?.watermark) {
      watermarkConf.value = config.watermark;
    }

    console.log(config);
  } catch (error) {
    console.error('Error load config:', error);
  }
};

onMounted(() => {
  loadConfig();
});

const prefixOptions = [
  { label: 'DAY(2006/01/02)', value: 'DAY' },
  { label: 'HOUR(2006/01/02/15/04)', value: 'HOUR' },
  { label: 'SECOND(2006/01/02/15/04/05)', value: 'SECOND' },
];

</script>

<template>
  <n-flex vertical>
    <n-card>
      <n-flex align="center">
        <label>类型: </label>
        <n-flex>
          <n-select style="width: 200px;" v-model:value="type"
            :options="[{ label: 'github', value: 'github' }]" />
        </n-flex>
      </n-flex>

      <n-flex>
        <n-card title="Github配置:">
          <n-grid cols="2">
            <n-gi>
              <label>token: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.token" placeholder="token" />
            </n-gi>

            <n-gi>
              <label>repo: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.repo" placeholder="repo" />
            </n-gi>

            <n-gi>
              <label>owner: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.owner" placeholder="owner" />
            </n-gi>

            <n-gi>
              <label>branch: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.branch" placeholder="branch" />
            </n-gi>

            <n-gi>
              <label>prefix: </label>
            </n-gi>
            <n-gi>
              <n-select v-model:value="githubConf.prefix" :options="prefixOptions" />
            </n-gi>

            <n-gi>
              <label>commit: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.commit" placeholder="commit" />
            </n-gi>

            <n-gi>
              <label>author name: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.author.name" placeholder="author name" />
            </n-gi>
            <n-gi>
              <label>author mail: </label>
            </n-gi>
            <n-gi>
              <n-input v-model:value="githubConf.author.mail" placeholder="author mail" />
            </n-gi>
          </n-grid>
        </n-card>

        <n-card title="水印:">
          <n-flex justify="left" align="center">
            <n-flex>
              <n-input :disabled="previewBlock" v-model:value="watermarkConf.text" placeholder="文本" @change="preview" />
            </n-flex>
            <n-flex>
              <n-input :disabled="previewBlock" v-model:value="watermarkConf.color" placeholder="颜色"
                @change="preview" />
            </n-flex>
            <n-flex>
              <n-input-number :disabled="previewBlock" v-model:value="watermarkConf.size" :precision="2"
                placeholder="大小" @change="preview" />
            </n-flex>
            <n-flex>
              <n-input-number :disabled="previewBlock" v-model:value="watermarkConf.dpi" :precision="2"
                placeholder="dpi" @change="preview" />
            </n-flex>
            <n-flex>
              <n-input-number :disabled="previewBlock" v-model:value="watermarkConf.x" placeholder="x"
                @change="preview" />
            </n-flex>
            <n-flex>
              <n-input-number :disabled="previewBlock" v-model:value="watermarkConf.y" placeholder="y"
                @change="preview" />
            </n-flex>
          </n-flex>
        </n-card>
      </n-flex>

      <template #action>

        <n-flex justify="center">
          <!-- <n-button type="info" @click="loadConfig">加载配置</n-button> -->
          <n-button type="primary" @click="saveConfig">保存配置</n-button>
        </n-flex>
      </template>
    </n-card>

    <n-divider />

    <n-flex justify="center" align="center">
      <n-button @click="handleFileSelect">选择文件</n-button>
    </n-flex>

    <n-flex vertical v-if="selectedFile">
      <n-alert title="选择的文件" type="info">
        {{ selectedFile }}
      </n-alert>

      <n-image width="100%" height="400" :object-fit="'contain'" :src="previewFile" alt="生成中..." />
      <n-flex justify="center">
        <n-button color="blue" @click="uploadFile">上传文件</n-button>
      </n-flex>
    </n-flex>

    <n-flex v-else>
      未选择文件
    </n-flex>

    <n-divider />

    <n-flex vertical v-if="uploadedUrl">
      <n-alert title="上传的文件" type="info">
        {{ uploadedUrl }}
        <a :href="uploadedUrl" target="_blank">点击访问原图</a>
      </n-alert>
      <n-image width="100%" height="400" :object-fit="'contain'" :src="uploadedUrl" alt="result" />
    </n-flex>
  </n-flex>

</template>

<style scoped></style>
